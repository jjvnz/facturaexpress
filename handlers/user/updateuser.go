package handlers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidID, "ID inválido"))
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrJSONBindingFailed, "Error al procesar los datos del usuario."))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrPasswordHashingFailed, "Error al hashear la contraseña."))
		return
	}

	db := data.GetInstance()

	query := "UPDATE usuarios SET nombre_usuario=$1, password=$2, correo=$3 WHERE id=$4"
	result, err := db.Exec(query, &user.Username, hashedPassword, &user.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	count, err := result.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrNotFound, "No se encontró el usuario con el ID especificado."))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Los datos del usuario se han actualizado correctamente."})
}
