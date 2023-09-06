package handlers

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListInvoices(c *gin.Context) {
	// Obtener el rol del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if rol != common.ADMIN && rol != common.USER {
		c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrInsuficientRole, "Lo siento, pero parece que no tienes los permisos necesarios para acceder a esta página. Por favor, verifica tus credenciales o contacta al ADMIN para obtener más información."))
		return
	}

	// Obtener y validar los parámetros de la consulta
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidPageParam, "El parámetro 'page' debe ser un número entero positivo"))
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidLimitParam, "El parámetro 'limit' debe ser un número entero positivo"))
		return
	}
	if limit > 100 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrLimitTooHigh, "El parámetro 'limit' no puede ser mayor a 100"))
		return
	}

	// Calcular el desplazamiento y los filtros de la consulta
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	filterField := c.Query("filter_field")
	filterValue := c.Query("filter_value")

	// Construir y ejecutar la consulta para obtener las facturas
	var rows *sql.Rows
	var query string
	var args []interface{}
	if filterField != "" && filterValue != "" {
		query = fmt.Sprintf(`SELECT * FROM facturas WHERE %s = $1 ORDER BY id ASC LIMIT $2 OFFSET $3`, filterField)
		args = []interface{}{filterValue, limit, offset}
	} else {
		if rol == common.ADMIN {
			query = `SELECT * FROM facturas ORDER BY id ASC LIMIT $1 OFFSET $2`
			args = []interface{}{limit, offset}
		} else {
			query = `SELECT * FROM facturas WHERE usuario_id = $1 ORDER BY id ASC LIMIT $2 OFFSET $3`
			args = []interface{}{claims.UserID, limit, offset}
		}
	}

	db := data.GetInstance()

	rows, err = db.Query(query, args...)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al obtener las facturas"))
		c.Abort()
		return
	}
	defer rows.Close()

	// Procesar las filas y construir el arreglo de facturas
	var invoices []models.Invoice
	for rows.Next() {
		var (
			id, userID                                    int
			companyName, companyTIN                       string
			date                                          string
			services                                      []byte
			totalValue                                    float64
			operatorName, documentType                    string
			document, documentIssuanceCity                string
			cellphone, bankAccountNumber, bankAccountType string
			bank                                          string
		)

		err := rows.Scan(&id, &companyName, &companyTIN, &date, &services, &totalValue, &operatorName, &documentType, &document, &documentIssuanceCity, &cellphone, &bankAccountNumber, &bankAccountType, &bank, &userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Ocurrió un error al leer los datos de la base de datos"))
			return
		}

		servicesDeserialized, err := helpers.UnmarshalServices(services)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrServicesUnMarshalError, "Error al deserializar servicios"))
			c.Abort()
			return
		}

		invoice := models.Invoice{
			ID:         id,
			Company:    models.Company{Name: companyName, TIN: companyTIN},
			Date:       helpers.FormatDateInSpanish(date),
			Services:   servicesDeserialized,
			TotalValue: totalValue,
			Operator:   models.Operator{Name: operatorName, DocumentType: documentType, Document: document, DocumentIssuanceCity: documentIssuanceCity, Cellphone: cellphone, BankAccountNumber: bankAccountNumber, BankAccountType: bankAccountType, Bank: bank},
			UserID:     int64(userID),
		}
		invoices = append(invoices, invoice)
	}

	// Verificar si se encontraron facturas y contar el total de facturas
	switch len(invoices) {
	case 0:
		c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrNotFound, "La página solicitada no existe"))
		c.Abort()
	default:
		var totalInvoices int
		if filterField != "" && filterValue != "" {
			query = `SELECT COUNT(*) FROM facturas WHERE $1 = $2`
			args = []interface{}{filterField, filterValue}
		} else {
			if rol == common.ADMIN {
				query = `SELECT COUNT(*) FROM facturas`
				args = []interface{}{}
			} else {
				query = `SELECT COUNT(*) FROM facturas WHERE usuario_id = $1`
				args = []interface{}{claims.UserID}
			}
		}
		err = db.QueryRow(query, args...).Scan(&totalInvoices)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al contar las facturas"))
			c.Abort()
			return
		}
		totalPages := int(math.Ceil(float64(totalInvoices) / float64(limit)))
		c.JSON(http.StatusOK, gin.H{
			"invoices":    invoices,
			"total_pages": totalPages,
			"page":        page,
		})
	}

}
