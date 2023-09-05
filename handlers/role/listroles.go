package handlers

import (
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListRoles(c *gin.Context) {
	db := data.GetInstance()

	rows, err := db.Query(`SELECT id, name FROM roles`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la lista de roles"})
		return
	}
	defer rows.Close()
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al escanear la fila de la base de datos"})
			return
		}
		roles = append(roles, role)
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}
