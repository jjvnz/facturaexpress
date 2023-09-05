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

func CreateInvoice(c *gin.Context) {
	// Obtener el rol y el ID de usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID

	// Comprobar si el usuario tiene el rol necesario para acceder a la ruta
	if !helpers.VerifyRole(role, []string{common.ADMIN, common.USER}) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrNoPermission, "No tienes permiso para acceder a esta página."))
		return
	}

	var invoice models.Invoice
	if err := c.BindJSON(&invoice); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit(common.ErrInvalidData, "Datos inválidos. Verifica y vuelve a intentarlo."))
		return
	}

	servicesJSON, err := json.Marshal(invoice.Services)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrServicesMarshalError, "Error al codificar los servicios en formato JSON."))
		return
	}

	db := data.GetInstance()

	query := `INSERT INTO facturas (nombre_empresa, nit_empresa, fecha, servicios, valor_total, nombre_operador, tipo_documento_operador, documento_operador, ciudad_expedicion_documento_operador, celular_operador, numero_cuenta_bancaria_operador, tipo_cuenta_bancaria_operador, banco_operador, usuario_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,$14) RETURNING id`
	err = db.QueryRow(query,
		invoice.Company.Name,
		invoice.Company.TIN,
		invoice.Date,
		servicesJSON,
		invoice.TotalValue,
		invoice.Operator.Name,
		invoice.Operator.DocumentType,
		invoice.Operator.Document,
		invoice.Operator.DocumentIssuanceCity,
		invoice.Operator.Cellphone,
		invoice.Operator.BankAccountNumber,
		invoice.Operator.BankAccountType,
		invoice.Operator.Bank,
		userID).Scan(&invoice.ID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrDBError, "Error al procesar las facturas"))
		c.Abort()
		return
	}

	// Actualiza el objeto de factura con el ID de usuario correcto
	invoice.UserID = userID

	c.JSON(http.StatusCreated, gin.H{"message": "Factura creada correctamente", "invoice": invoice})
}
