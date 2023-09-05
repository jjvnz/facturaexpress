package handlers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidID, "ID inválido"))
		return
	}

	db := data.GetInstance()

	query := "DELETE FROM usuarios WHERE id = $1"
	result, err := db.Exec(query, id)
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
	c.JSON(http.StatusOK, gin.H{"message": "El registro ha sido eliminado correctamente."})
}
