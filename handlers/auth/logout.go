package handlers

import (
	"facturaexpress/data"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// En tu archivo handlers.go, agrega una nueva función para manejar solicitudes de logout
func Logout(c *gin.Context, db *data.DB) {
	/// En tu función Logout, después de obtener el token del encabezado Authorization:
	tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	// Verificar si el token ya está en la lista negra
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM jwt_blacklist WHERE token = $1`, tokenString).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el token está en la lista negra"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El token ya está en la lista negra"})
		return
	}

	// Agregar el token a la lista negra para revocarlo
	stmt, err := db.Prepare(`INSERT INTO jwt_blacklist (token) VALUES ($1)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al agregar el token a la lista negra"})
		return
	}

	// Devolver una respuesta al cliente indicando que la sesión ha sido cerrada con éxito
	c.JSON(http.StatusOK, gin.H{
		"message": "Sesión cerrada con éxito",
	})
}
