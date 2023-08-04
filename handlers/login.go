package handlers

import (
	"database/sql"
	"facturaexpress/data"
	"facturaexpress/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Login handles user login and token generation.
func Login(c *gin.Context, db *data.DB, jwtKey []byte) {
	var loginData models.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		errorResponse := models.ErrorResponseInit("BAD_REQUEST", "Error al leer los datos de inicio de sesión.")
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	user, err := verifyCredentials(db, loginData.Correo, loginData.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse := models.ErrorResponseInit("EMAIL_NOT_FOUND", "No se encontró ningún usuario con el correo electrónico que ingresaste")
			c.JSON(http.StatusUnauthorized, errorResponse)
		} else {
			errorResponse := models.ErrorResponseInit("INCORRECT_PASSWORD", "La contraseña que ingresaste es incorrecta")
			c.JSON(http.StatusUnauthorized, errorResponse)
		}
		return
	}

	tokenString, err := generateJWTToken(jwtKey, user.ID, user.Role)
	if err != nil {
		log.Printf("%v", err)
		errorResponse := models.ErrorResponseInit("JWT_GENERATION_ERROR", "No se pudo generar el token JWT debido a un problema interno")
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Declarar la variable stmt antes de utilizarla
	var stmt *sql.Stmt

	// En tu función Login, después de generar el token JWT:
	stmt, err = db.Prepare(`UPDATE usuarios SET jwt_token = $1 WHERE id = $2`)
	if err != nil {
		errorResponse := models.ErrorResponseInit("JWT_STORAGE_ERROR", "Ocurrió un problema al intentar almacenar el token JWT del usuario en la base de datos")
		c.JSON(http.StatusInternalServerError, errorResponse)
		return // Agregar esta línea para detener la ejecución del código si ocurre un error
	}
	defer stmt.Close()
	_, err = stmt.Exec(tokenString, user.ID)
	if err != nil {
		errorResponse := models.ErrorResponseInit("JWT_STORAGE_ERROR", "Ocurrió un problema al intentar almacenar el token JWT del usuario en la base de datos")
		c.JSON(http.StatusInternalServerError, errorResponse)
		return // Agregar esta línea para detener la ejecución del código si ocurre un error
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inicio de sesión exitoso", "token": tokenString})
}

// verifyCredentials verifies the user's email and password.
func verifyCredentials(db *data.DB, correo string, password string) (models.Usuario, error) {
	var user models.Usuario
	stmt, err := db.Prepare(`SELECT usuarios.id, usuarios.nombre_usuario, usuarios.password, roles.name
		FROM usuarios
		INNER JOIN user_roles ON usuarios.id = user_roles.user_id
		INNER JOIN roles ON user_roles.role_id = roles.id
		WHERE usuarios.correo=$1`)
	if err != nil {
		return user, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(correo)
	err = row.Scan(&user.ID, &user.NombreUsuario, &user.Password, &user.Role)
	if err == sql.ErrNoRows {
		return user, err
	} else if err != nil {
		return user, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return user, models.ErrorResponseInit("INCORRECT_PASSWORD", "La contraseña que ingresaste es incorrecta.")
	}

	return user, nil
}

// generateJWTToken generates a JWT token for the given user ID.
func generateJWTToken(jwtKey []byte, usuarioID int64, role string) (string, error) {
	expTimeStr := os.Getenv("JWT_EXP_TIME")
	expDuration, _ := time.ParseDuration(expTimeStr)
	if expDuration == 0 {
		expDuration = 24 * time.Hour //default value of 24 hours if not set in env variable or if there is an error parsing it.
	}
	expTime := time.Now().Add(expDuration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &models.Claims{
		UsuarioID: usuarioID,
		Role:      role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expTime.Unix(),
		},
	})
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", models.ErrorResponseInit("ERROR_GENERATING_TOKEN", "Error al generar el token JWT.")
	}
	return tokenString, nil
}
