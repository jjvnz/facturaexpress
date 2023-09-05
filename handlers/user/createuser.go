package handlers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrJSONBindingFailed, "Error al procesar los datos del usuario."))
		return
	}

	db := data.GetInstance()

	if err := helpers.CheckUsernameEmail(db, user.Username, user.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrPasswordHashingFailed, "Error al hashear la contraseña."))
		return
	}

	query := "INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id"
	err = db.QueryRow(query, user.Username, hashedPassword, user.Email).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryFailed, "Error al ejecutar la consulta."))
		return
	}

	c.JSON(http.StatusCreated, user)
}
