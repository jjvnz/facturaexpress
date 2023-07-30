package handlers

import (
	"facturaexpress/pkg/models"
	"facturaexpress/pkg/storage"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context, db *storage.DB) {
	var user models.Usuario
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar si el nombre de usuario ya existe
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM usuarios WHERE nombre_usuario = $1", user.NombreUsuario)
	err := row.Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el nombre de usuario"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre de usuario ya está en uso"})
		return
	}

	// Verificar si el correo electrónico ya existe
	row = db.QueryRow("SELECT COUNT(*) FROM usuarios WHERE correo = $1", user.Correo)
	err = row.Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el correo electrónico"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El correo electrónico ya está en uso"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al hashear la contraseña"})
		return
	}

	query := `INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3)`
	_, err = db.Exec(query, user.NombreUsuario, hashedPassword, user.Correo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el usuario en la base dEe datos"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado con éxito"})
}
