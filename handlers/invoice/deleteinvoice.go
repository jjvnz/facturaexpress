package handlers

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DeleteInvoice(c *gin.Context) {
	// Get the role and user ID from the JWT token
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID

	// Check if the user has the necessary role to access the route
	allowedRoles := []string{common.ADMIN, common.USER}
	if !helpers.VerifyRole(role, allowedRoles) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrNoPermission, "No tienes permiso para acceder a esta página."))
		c.Abort()
		return
	}

	// Validate the value of the id parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidID, "El valor del parámetro id no es un número válido"))
		c.Abort()
		return
	}
	if id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidID, "El valor del parámetro id debe ser un número entero positivo"))
		c.Abort()
		return
	}

	var query string
	if role == common.ADMIN {
		query = `DELETE FROM facturas WHERE id = $1`
	} else {
		query = `DELETE FROM facturas WHERE id = $1 AND usuario_id = $2`
	}

	var result sql.Result

	db := data.GetInstance()

	if role == common.ADMIN {
		result, err = db.Exec(query, id)
	} else {
		result, err = db.Exec(query, id, userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrQueryFailed, "Error al ejecutar la consulta SQL"))
		c.Abort()
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura eliminada correctamente"})
	} else {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrInvoiceNotFound, "No se encontró la factura con el ID especificado o no tienes permiso para eliminarla"))
		c.Abort()
	}
}
