package handlers

import (
	"encoding/json"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UpdateInvoice(c *gin.Context) {
	// Get the role and user ID from the JWT token
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID
	invoiceID := c.Param("id")

	// Check if the user has the necessary role to access the route
	if !helpers.VerifyRole(role, []string{common.ADMIN, common.USER}) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrNoPermission, "No tienes permiso para acceder a esta página."))
		c.Abort()
		return
	}

	// Add a condition to allow common.ADMIN role to update any invoice
	if role != common.ADMIN {
		// Check if the user is trying to update their own invoice
		invoiceUserID, err := helpers.GetUserIDFromInvoice(invoiceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al obtener el ID del usuario de la factura."))
			c.Abort()
			return
		}
		if userID != invoiceUserID {
			c.JSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrNoPermission, "Solo puedes actualizar tus propias facturas."))
			c.Abort()
			return
		}
	}

	var invoice models.Invoice
	err := c.BindJSON(&invoice)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidData, "Datos inválidos. Verifica y vuelve a intentarlo."))
		c.Abort()
		return
	}

	// Validate input data
	if invoice.Company.Name == "" || invoice.Company.TIN == "" || invoice.Date.IsZero() || len(invoice.Services) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrMissingFields, "Faltan campos requeridos."))
		c.Abort()
		return
	}

	servicesJSON, err := json.Marshal(invoice.Services)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrServicesMarshalError, "Error al codificar los servicios en formato JSON."))
		c.Abort()
		return
	}

	// Check if the user exists
	var userExists bool
	db := data.GetInstance()
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al verificar si el usuario existe."))
		c.Abort()
		return
	}
	if !userExists {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrUserNotFound, "El usuario especificado no existe."))
		c.Abort()
		return
	}

	query := `UPDATE facturas SET nombre_empresa = $1,nit_empresa = $2,
			fecha = $3,servicios = $4,
			valor_total = $5,nombre_operador = $6,
			tipo_documento_operador = $7,
			documento_operador = $8,
			ciudad_expedicion_documento_operador = $9,
			celular_operador = $10,
			numero_cuenta_bancaria_operador = $11,
			tipo_cuenta_bancaria_operador = $12,
			banco_operador = $13 WHERE id = $14`
	result, err := db.Exec(query, invoice.Company.Name, invoice.Company.TIN, invoice.Date, servicesJSON, invoice.TotalValue, invoice.Operator.Name, invoice.Operator.DocumentType, invoice.Operator.Document, invoice.Operator.DocumentIssuanceCity, invoice.Operator.Cellphone, invoice.Operator.BankAccountNumber, invoice.Operator.BankAccountType, invoice.Operator.Bank, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al actualizar la factura en la base de datos."))
		c.Abort()
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura actualizada correctamente"})
	} else {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrInvoiceNotFound, "No se encontró la factura con el ID especificado"))
	}
}
