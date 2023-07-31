package handlers

import (
	"database/sql"
	"facturaexpress/pkg/models"
	"facturaexpress/pkg/storage"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Login handles user login and token generation.
func Login(c *gin.Context, db *storage.DB, jwtKey []byte) {
	var loginData models.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		sendResponse(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := verifyCredentials(db, loginData.Correo, loginData.Password)
	if err != nil {
		sendResponse(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := generateJWTToken(jwtKey, user.ID)
	if err != nil {
		log.Printf("%v", err)
		sendResponse(c, http.StatusInternalServerError, gin.H{"error": "Error al generar el token JWT"})
		return
	}

	sendResponse(c, http.StatusOK, gin.H{"message": "Inicio de sesión exitoso", "token": tokenString})
}

// verifyCredentials verifies the user's email and password.
func verifyCredentials(db *storage.DB, correo string, password string) (models.Usuario, error) {
	var user models.Usuario
	row := db.QueryRow("SELECT id, nombre_usuario, password FROM usuarios WHERE correo = $1", correo)
	err := row.Scan(&user.ID, &user.NombreUsuario, &user.Password)
	if err == sql.ErrNoRows {
		return user, fmt.Errorf("credenciales inválidas")
	} else if err != nil {
		return user, fmt.Errorf("error al verificar las credenciales")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return user, fmt.Errorf("credenciales inválidas")
	}

	return user, nil
}

// generateJWTToken generates a JWT token for the given user ID.
func generateJWTToken(jwtKey []byte, usuarioID int64) (string, error) {
	expTimeStr := os.Getenv("JWT_EXP_TIME")
	expDuration, err := time.ParseDuration(expTimeStr)
	if err != nil {
		return "", fmt.Errorf("error al analizar la duración del tiempo de expiración del JWT: %v", err)
	}
	expTime := time.Now().Add(expDuration)
	claims := &models.Claims{
		UsuarioID: usuarioID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("error al generar el token JWT: %v", err)
	}

	return tokenString, nil
}

// sendResponse sends a JSON response with the given status code and data.
func sendResponse(c *gin.Context, statusCode int, data gin.H) {
	c.JSON(statusCode, data)
}
