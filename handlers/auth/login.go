package handlers

import (
	"database/sql"
	"facturaexpress/data"
	"facturaexpress/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Login maneja el inicio de sesión del usuario y la generación de tokens.
func Login(c *gin.Context, db *data.DB, jwtKey []byte, expTimeStr string) {
	var loginData models.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("BAD_REQUEST", "Error al leer los datos de inicio de sesión."))
		return
	}

	user, err := verifyCredentials(db, loginData.Email, loginData.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseInit("EMAIL_NOT_FOUND", "No se encontró ningún usuario con el correo electrónico que ingresaste"))
		} else {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseInit("INCORRECT_PASSWORD", "La contraseña que ingresaste es incorrecta."))
		}
		return
	}

	tokenString, err := generateJWTToken(jwtKey, user.ID, user.Role, expTimeStr)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("JWT_GENERATION_ERROR", "No se pudo generar el token JWT debido a un problema interno"))
		return
	}

	stmt, err := db.Prepare(`UPDATE usuarios SET jwt_token = $1 WHERE id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("JWT_STORAGE_ERROR", "Ocurrió un problema al intentar almacenar el token JWT del usuario en la base de datos"))
		return
	}
	defer stmt.Close()
	if _, err = stmt.Exec(tokenString, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("JWT_STORAGE_ERROR", "Ocurrió un problema al intentar almacenar el token JWT del usuario en la base de datos"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inicio de sesión exitoso", "token": tokenString})
}

// verifyCredentials verifica el correo electrónico y la contraseña del usuario.
func verifyCredentials(db *data.DB, correo string, password string) (models.User, error) {
	var user models.User
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
	err = row.Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err == sql.ErrNoRows {
		return user, err
	} else if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return user, models.ErrorResponseInit("INCORRECT_PASSWORD", "La contraseña que ingresaste es incorrecta.")
	}
	return user, nil
}

// generateJWTToken genera un token JWT para el ID de usuario dado.
func generateJWTToken(jwtKey []byte, userID int64, role string, expTimeStr string) (string, error) {
	expDuration := 24 * time.Hour // valor predeterminado de 24 horas si no se establece en la variable expTimeStr o si hay un error al analizarlo.
	if d, _ := time.ParseDuration(expTimeStr); d > 0 {
		expDuration = d
	}
	expTime := time.Now().Add(expDuration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &models.Claims{
		UserID: userID,
		Role:   role,
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
