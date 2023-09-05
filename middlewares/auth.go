package middleware

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context, jwtKey []byte) {
	claims, errCode, err := helpers.VerifyToken(c, jwtKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponseInit(errCode, err.Error()))
		c.Abort()
		return
	}

	db := data.GetInstance()

	tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM jwt_blacklist WHERE token = $1`)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al preparar la consulta"))
		c.Abort()
		return
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(tokenString).Scan(&count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al verificar si el token está en la lista negra"))
		c.Abort()
		return
	}
	if count > 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponseInit(common.ErrInvalidToken, "Token inválido. Verifica o solicita uno nuevo."))
		c.Abort()
		return
	}

	c.Set("claims", claims)

	c.Next()
}

func RoleAuthMiddleware(c *gin.Context, role string) {
	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	claims := c.MustGet("claims").(*models.Claims)
	userID := claims.UserID

	// Agregar una condición para permitir que el rol de ADMIN acceda a la ruta
	if claims.Role == common.ADMIN {
		c.Next()
		return
	}

	db := data.GetInstance()

	var userRole models.UserRole
	stmt, err := db.Prepare(`SELECT user_id, role_id FROM user_roles JOIN roles ON user_roles.role_id = roles.id WHERE user_roles.user_id = $1 AND roles.name = $2`)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al preparar la consulta"))
		c.Abort()
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userID, role).Scan(&userRole.UserID, &userRole.RoleID)
	if err == sql.ErrNoRows {
		c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrInsuficientRole, "No tiene el rol necesario para acceder a esta ruta"))
		c.Abort()
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al verificar si el usuario tiene el rol necesario"))
		c.Abort()
		return
	}
	c.Next()
}
