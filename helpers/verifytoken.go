package helpers

import (
	"facturaexpress/common"
	"facturaexpress/models"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func VerifyToken(c *gin.Context, jwtKey []byte) (*models.Claims, string, error) {
	// Extrae el token JWT del encabezado Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, common.ErrInvalidAuthHeader, fmt.Errorf("falta encabezado Authorization o prefijo Bearer")
	}

	// Elimina el prefijo Bearer y el espacio del encabezado Authorization
	jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Verifica la firma y valida los claims del token JWT
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, common.ErrInvalidToken, fmt.Errorf("token inválido Verifica o solicita uno nuevo")
	}

	return claims, "", nil
}
