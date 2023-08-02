package middleware

import (
	"database/sql"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(c *gin.Context, db *data.DB, jwtKey []byte) {
	// Extrae el token JWT del encabezado Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		errorResponse := models.ErrorResponseInit("INVALID_AUTH_HEADER", "Falta encabezado Authorization o prefijo Bearer")
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse)
		c.Abort()
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
		errorResponse := models.ErrorResponseInit("INVALID_TOKEN", "Token inválido. Verifica o solicita uno nuevo.")
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse)
		c.Abort()
		return
	}

	// En tu middleware AuthMiddleware, después de verificar la firma y validar los claims del token JWT:
	tokenString := c.GetHeader("Authorization")
	if tokenString != "" {
		stmt, err := db.Prepare(`SELECT COUNT(*) FROM jwt_blacklist WHERE token = $1`)
		if err != nil {
			errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al preparar la consulta")
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			c.Abort()
			return
		}
		defer stmt.Close()
		var count int
		err = stmt.QueryRow(tokenString).Scan(&count)
		if err != nil {
			errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al verificar si el token está en la lista negra")
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			c.Abort()
			return
		}
		if count > 0 {
			errorResponse := models.ErrorResponseInit("INVALID_TOKEN", "Token inválido. Verifica o solicita uno nuevo.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse)
			c.Abort()
			return
		}
	}

	// Establece los claims en el contexto de Gin
	//c.Set("usuario_id", claims.UsuarioID)
	c.Set("claims", claims)

	c.Next()
}

func RoleAuthMiddleware(c *gin.Context, db *data.DB, role string) {
	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	claims := c.MustGet("claims").(*models.Claims)
	userID := claims.UsuarioID

	var userRole models.UserRole
	stmt, err := db.Prepare(`SELECT user_id, role_id FROM user_roles JOIN roles ON user_roles.role_id = roles.id WHERE user_roles.user_id = $1 AND roles.name = $2`)
	if err != nil {
		errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al preparar la consulta")
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userID, role).Scan(&userRole.UserID, &userRole.RoleID)
	if err == sql.ErrNoRows {
		errorResponse := models.ErrorResponseInit("INSUFFICIENT_ROLE", "No tiene el rol necesario para acceder a esta ruta")
		c.AbortWithStatusJSON(http.StatusForbidden, errorResponse)
		c.Abort()
		return
	} else if err != nil {
		errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al verificar si el usuario tiene el rol necesario")
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	c.Next()
}
