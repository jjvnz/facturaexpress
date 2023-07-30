package handlers

import (
	"database/sql"
	"facturaexpress/pkg/models"
	"facturaexpress/pkg/storage"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context, db *storage.DB, jwtKey []byte) {
	var loginData models.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.Usuario
	row := db.QueryRow("SELECT id, nombre_usuario, password FROM usuarios WHERE correo = $1", loginData.Correo)
	err := row.Scan(&user.ID, &user.NombreUsuario, &user.Password)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar las credenciales"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// Si las credenciales son válidas, genera un token JWT
	expTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		UsuarioID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Error al generar el token JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar el token JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inicio de sesión exitoso", "token": tokenString})
}
