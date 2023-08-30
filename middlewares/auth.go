package middleware

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(c *gin.Context, jwtKey []byte) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponseInit("INVALID_AUTH_HEADER", "Falta encabezado Authorization o prefijo Bearer"))
		c.Abort()
		return
	}

	jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponseInit("INVALID_TOKEN", "Token inválido. Verifica o solicita uno nuevo."))
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	db := data.GetInstance()
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM jwt_blacklist WHERE token = $1`)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al preparar la consulta"))
		c.Abort()
		return
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(tokenString).Scan(&count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al verificar si el token está en la lista negra"))
		c.Abort()
		return
	}
	if count > 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponseInit("INVALID_TOKEN", "Token inválido. Verifica o solicita uno nuevo."))
		c.Abort()
		return
	}

	c.Set("claims", claims)

	c.Next()
}

func RoleAuthMiddleware(c *gin.Context, role string) {
	claims := c.MustGet("claims").(*models.Claims)
	userID := claims.UserID

	if claims.Role == common.ADMIN {
		c.Next()
		return
	}

	db := data.GetInstance()

	var userRole models.UserRole
	stmt, err := db.Prepare(`SELECT user_id, role_id FROM user_roles JOIN roles ON user_roles.role_id = roles.id WHERE user_roles.user_id = $1 AND roles.name = $2`)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al preparar la consulta"))
		c.Abort()
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userID, role).Scan(&userRole.UserID, &userRole.RoleID)
	if err == sql.ErrNoRows {
		c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponseInit("INSUFFICIENT_ROLE", "No tiene el rol necesario para acceder a esta ruta"))
		c.Abort()
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al verificar si el usuario tiene el rol necesario"))
		c.Abort()
		return
	}
	c.Next()
}
