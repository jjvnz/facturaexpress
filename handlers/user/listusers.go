package handlers

import (
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListUsers(c *gin.Context) {
	db := data.GetInstance()

	rows, err := db.Query(`SELECT usuarios.id, usuarios.nombre_usuario, usuarios.password, usuarios.correo, roles.name
	FROM usuarios
	INNER JOIN user_roles ON usuarios.id = user_roles.user_id
	INNER JOIN roles ON user_roles.role_id = roles.id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_FAILED", "Error al ejecutar la consulta."))
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("SCAN_FAILED", "Error al escanear los resultados."))
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}
