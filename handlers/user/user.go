package handlers

import (
	"net/http"
	"strconv"

	"facturaexpress/data"
	handlers "facturaexpress/handlers/auth"
	"facturaexpress/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ListUsers(c *gin.Context, db *data.DB) {
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

func CreateUser(c *gin.Context, db *data.DB) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("JSON_BINDING_FAILED", "Error al procesar los datos del usuario."))
		return
	}

	if err := handlers.CheckUsernameEmail(db, user.Username, user.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("PASSWORD_HASHING_FAILED", "Error al hashear la contraseña."))
		return
	}

	query := "INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id"
	err = db.QueryRow(query, user.Username, hashedPassword, user.Email).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_FAILED", "Error al ejecutar la consulta."))
		return
	}

	c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context, db *data.DB) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_ID", "ID inválido"))
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("JSON_BINDING_FAILED", "Error al procesar los datos del usuario."))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("PASSWORD_HASHING_FAILED", "Error al hashear la contraseña."))
		return
	}

	query := "UPDATE usuarios SET nombre_usuario=$1, password=$2, correo=$3 WHERE id=$4"
	result, err := db.Exec(query, &user.Username, hashedPassword, &user.Email, id)
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
	c.JSON(http.StatusOK, gin.H{"message": "Los datos del usuario se han actualizado correctamente."})
}

func DeleteUser(c *gin.Context, db *data.DB) {
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
	c.JSON(http.StatusOK, gin.H{"message": "El registro ha sido eliminado correctamente."})
}
