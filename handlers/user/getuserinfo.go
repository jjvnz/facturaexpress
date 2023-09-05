package handlers

import (
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context) {
	db := data.GetInstance()

	// Obtener el ID del usuario autenticado y su rol
	claims := c.MustGet("claims").(*models.Claims)
	userID := claims.UserID
	//userRole := claims.Role

	/* Verificar si el rol del usuario autenticado es "usuario"
	if userRole != common.USER {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit("FORBIDDEN", "No tienes permiso para acceder a este recurso."))
		return
	}*/

	// Consultar la información del usuario autenticado
	row := db.QueryRow(`SELECT usuarios.id, usuarios.nombre_usuario, usuarios.correo, roles.name
	FROM usuarios
	INNER JOIN user_roles ON usuarios.id = user_roles.user_id
	INNER JOIN roles ON user_roles.role_id = roles.id
	WHERE usuarios.id = $1`, userID)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("SCAN_FAILED", "Error al escanear los resultados."))
		return
	}

	c.JSON(http.StatusOK, user)
}
