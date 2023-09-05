package handlers

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UpdateRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleID")

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidUserID, "El ID del usuario debe ser un número entero válido"))
		return
	}
	roleIDInt, err := strconv.Atoi(roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidRoleID, "El ID del nuevo rol debe ser un número entero válido"))
		return
	}

	db := data.GetInstance()

	var count int
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrUserVerificationFailed, "Error al verificar si el usuario existe"))
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrUserNotFound, "El usuario especificado no existe"))
		return
	}
	stmt, err = db.Prepare(`SELECT COUNT(*) FROM roles WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrRoleVerificationFailed, "Error al verificar si el rol existe"))
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrRoleNotFound, "El rol especificado no existe"))
		return
	}

	stmt, err = db.Prepare(`SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt, roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrUserRoleVerificationFailed, "Error al verificar si el usuario ya tiene el rol especificado"))
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrUserAlreadyHasRole, "El usuario ya tiene el rol especificado. Por favor, actualice a otro rol."))
		return
	}

	stmt, err = db.Prepare(`UPDATE user_roles SET role_id = $1 WHERE user_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(roleIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrUserRoleUpdateFailed, "Error al actualizar el rol del usuario"))
		return
	}

	stmt, err = db.Prepare(`SELECT jwt_token FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	var tokenString sql.NullString
	err = stmt.QueryRow(userIDInt).Scan(&tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			models.ErrorResponseInit(common.ErrJWTTokenRetrievalFailed,
				"Error al obtener el token JWT del usuario"))
		return
	}
	if tokenString.Valid && tokenString.String != "" {
		stmt, err := db.Prepare(`INSERT INTO jwt_blacklist (token) VALUES ($1)`)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponseInit(common.ErrQueryPreparationFailed,
					"Error al preparar la consulta"))
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(tokenString.String)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponseInit(common.ErrJWTTokenBlacklistingFailed,
					"Error al agregar el token a la lista negra"))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rol actualizado con éxito",
	})
}
