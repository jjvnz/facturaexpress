package handlers

import (
	"net/http"
	"strconv"

	"facturaexpress/data"
	"facturaexpress/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ListarUsuarios(c *gin.Context, db *data.DB) {
	rows, err := db.Query(`SELECT usuarios.id, usuarios.nombre_usuario, usuarios.password, usuarios.correo, roles.name
	FROM usuarios
	INNER JOIN user_roles ON usuarios.id = user_roles.user_id
	INNER JOIN roles ON user_roles.role_id = roles.id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_FAILED", "Error al ejecutar la consulta."))
		return
	}
	defer rows.Close()

	var usuarios []models.Usuario
	for rows.Next() {
		var usuario models.Usuario
		err = rows.Scan(&usuario.ID, &usuario.Nombre, &usuario.Password, &usuario.Correo, &usuario.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("SCAN_FAILED", "Error al escanear los resultados."))
			return
		}
		usuarios = append(usuarios, usuario)
	}

	c.JSON(http.StatusOK, usuarios)
}

func CrearUsuario(c *gin.Context, db *data.DB) {
	var usuario models.Usuario
	if err := c.ShouldBindJSON(&usuario); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("JSON_BINDING_FAILED", "Error al procesar los datos del usuario."))
		return
	}

	if err := checkUsernameEmail(db, usuario.Nombre, usuario.Correo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(usuario.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("PASSWORD_HASHING_FAILED", "Error al hashear la contraseña."))
		return
	}

	query := "INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id"
	err = db.QueryRow(query, usuario.Nombre, hashedPassword, usuario.Correo).Scan(&usuario.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_FAILED", "Error al ejecutar la consulta."))
		return
	}

	c.JSON(http.StatusCreated, usuario)
}

func ActualizarUsuario(c *gin.Context, db *data.DB) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_ID", "ID inválido"))
		return
	}

	var usuario models.Usuario
	if err := c.ShouldBindJSON(&usuario); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("JSON_BINDING_FAILED", "Error al procesar los datos del usuario."))
		return
	}

	query := "UPDATE usuarios SET nombre_usuario=$1,password=$2 WHERE id=$3"
	result, err := db.Exec(query, &usuario.Nombre, &usuario.Password, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	count, err := result.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, "No se encontró el registro")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "Registro actualizado")
}

func EliminarUsuario(c *gin.Context, db *data.DB) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_ID", "ID inválido"))
		return
	}

	query := "DELETE FROM usuarios WHERE id = $1"
	result, err := db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	count, err := result.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit("NOT_FOUND", "No se encontró el usuario con el ID especificado."))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "Registro eliminado")
}
