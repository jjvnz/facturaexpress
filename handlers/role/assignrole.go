package handlers

import (
	"facturaexpress/data"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AssignRole(c *gin.Context) {
	userID := c.Param("id")
	newRoleID := c.Param("newRoleID")

	if userID == "" || newRoleID == " " {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error en los paramentros de consulta"})
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del usuario debe ser un número entero válido"})
		return
	}
	roleIDInt, err := strconv.Atoi(newRoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del rol debe ser un número entero válido"})
		return
	}

	db := data.GetInstance()

	var count int
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el usuario existe"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El usuario especificado no existe"})
		return
	}
	stmt, err = db.Prepare(`SELECT COUNT(*) FROM roles WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el rol existe"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El rol especificado no existe"})
		return
	}

	stmt, err = db.Prepare(`SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt, roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el usuario ya tiene asignado el rol especificado"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El usuario ya tiene asignado el rol especificado"})
		return
	}

	stmt, err = db.Prepare(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(userIDInt, roleIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al asignar el rol al usuario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rol asignado con éxito",
	})
}
