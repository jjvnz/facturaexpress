package middleware

import (
	"facturaexpress/pkg/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(c *gin.Context, jwtKey []byte) {
	// Extrae el token JWT del encabezado Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Falta encabezado Authorization o prefijo Bearer"})
		return
	}

	// Elimina el prefijo Bearer y el espacio del encabezado Authorization
	jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Verifica la firma y valida los claims del token JWT
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// Establece el ID del usuario en el contexto de Gin
	c.Set("usuario_id", claims.UsuarioID)

	c.Next()
}
