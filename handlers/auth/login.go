package handlers

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login maneja el inicio de sesión del usuario y la generación de tokens.
func Login(c *gin.Context, jwtKey []byte, expTimeStr string) {
	var loginData models.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrBadRequest, "Error al leer los datos de inicio de sesión."))
		return
	}

	db := data.GetInstance()

	user, err := helpers.VerifyCredentials(db, loginData.Email, loginData.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseInit(common.ErrEmailNotFound, "No se encontró ningún usuario con el correo electrónico que ingresaste"))
		} else {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseInit(common.ErrIncorrectPassword, "La contraseña que ingresaste es incorrecta."))
		}
		return
	}

	tokenString, err := helpers.GenerateJWTToken(jwtKey, user.ID, user.Role, expTimeStr)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrJWTGenerationError, "No se pudo generar el token JWT debido a un problema interno"))
		return
	}

	stmt, err := db.Prepare(`UPDATE usuarios SET jwt_token = $1 WHERE id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrJWTStorageError, "Ocurrió un problema al intentar almacenar el token JWT del usuario en la base de datos"))
		return
	}
	defer stmt.Close()
	if _, err = stmt.Exec(tokenString, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrJWTStorageError, "Ocurrió un problema al intentar almacenar el token JWT del usuario en la base de datos"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inicio de sesión exitoso", "token": tokenString})
}
