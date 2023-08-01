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

	stmt, err := db.Prepare("SELECT COUNT(*) FROM usuarios WHERE nombre_usuario = $1 OR correo = $2")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow(user.NombreUsuario, user.Correo)
	var count int
	err = row.Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el nombre de usuario y el correo electrónico"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre de usuario o el correo electrónico ya están en uso"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al hashear la contraseña"})
		return
	}

	stmt, err = db.Prepare(`INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	var userID int64
	err = stmt.QueryRow(user.NombreUsuario, hashedPassword, user.Correo).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el usuario en la base de datos"})
		return
	}

	stmt, err = db.Prepare(`SELECT id FROM roles WHERE name=$1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	row = stmt.QueryRow("usuario")
	var roleID int64
	err = row.Scan(&roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el ID del rol"})
		return
	}

	stmt, err = db.Prepare(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al asignar el rol al usuario"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado con éxito"})
}
