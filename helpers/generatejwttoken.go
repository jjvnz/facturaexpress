// middleware/jwt/jwt.go
package helpers

import (
	"facturaexpress/common"
	"facturaexpress/models"
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateJWTToken(jwtKey []byte, userID int64, role string, expTimeStr string) (string, error) {
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
		return "", models.ErrorResponseInit(common.ErrErrorGeneratingToken, "Error al generar el token JWT.")
	}
	return tokenString, nil
}
