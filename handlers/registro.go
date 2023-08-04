package handlers

import (
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context, db *data.DB) {
	var user models.Usuario
	if err := c.ShouldBindJSON(&user); err != nil {
		errorResponse := models.ErrorResponseInit("JSON_BINDING_FAILED", "Error al procesar los datos del usuario.")
		c.JSON(http.StatusBadRequest, errorResponse)
		c.Abort()
		return
	}

	stmt, err := db.Prepare("SELECT COUNT(*) FROM usuarios WHERE nombre_usuario = $1 OR correo = $2")
	if err != nil {
		errorResponse := models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow(user.NombreUsuario, user.Correo)
	var count int
	err = row.Scan(&count)
	if err != nil {
		errorResponse := models.ErrorResponseInit("USERNAME_EMAIL_VERIFICATION_FAILED", "Error al verificar el nombre de usuario y el correo electrónico.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	if count > 0 {
		errorResponse := models.ErrorResponseInit("USERNAME_OR_EMAIL_IN_USE", "El nombre de usuario o el correo electrónico ya están en uso.")
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse := models.ErrorResponseInit("PASSWORD_HASHING_FAILED", "Error al hashear la contraseña.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	stmt, err = db.Prepare(`INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id`)
	if err != nil {
		errorResponse := models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	defer stmt.Close()
	var userID int64
	err = stmt.QueryRow(user.NombreUsuario, hashedPassword, user.Correo).Scan(&userID)
	if err != nil {
		errorResponse := models.ErrorResponseInit("DATABASE_SAVE_FAILED", "Error al guardar en la base de datos.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	stmt, err = db.Prepare(`SELECT id FROM roles WHERE name=$1`)
	if err != nil {
		errorResponse := models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	defer stmt.Close()
	row = stmt.QueryRow("usuario")
	var roleID int64
	err = row.Scan(&roleID)
	if err != nil {
		errorResponse := models.ErrorResponseInit("ROLE_ID_RETRIEVAL_FAILED", "Error al obtener el ID del rol.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	stmt, err = db.Prepare(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`)
	if err != nil {
		errorResponse := models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID, roleID)
	if err != nil {
		errorResponse := models.ErrorResponseInit("DATABASE_SAVE_FAILED", "Error al guardar en la base de datos.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado con éxito."})
}
