package handlers

import (
	"database/sql"
	"encoding/json"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func unmarshalServices(data []byte) ([]models.Service, error) {
	var services []models.Service
	err := json.Unmarshal(data, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func ListInvoices(c *gin.Context) {

	// Obtener el rol del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if rol != common.ADMIN && rol != common.USER {
		c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponseInit("INSUFFICIENT_ROLE", "Lo siento, pero parece que no tienes los permisos necesarios para acceder a esta página. Por favor, verifica tus credenciales o contacta al ADMIN para obtener más información."))
		return
	}

	// Obtener y validar los parámetros de la consulta
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_PAGE_PARAM", "El parámetro 'page' debe ser un número entero positivo"))
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_LIMIT_PARAM", "El parámetro 'limit' debe ser un número entero positivo"))
		return
	}
	if limit > 100 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseInit("LIMIT_TOO_HIGH", "El parámetro 'limit' no puede ser mayor a 100"))
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al obtener las facturas"))
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
			date                                          time.Time
			services                                      []byte
			totalValue                                    float64
			operatorName, documentType                    string
			document, documentIssuanceCity                string
			cellphone, bankAccountNumber, bankAccountType string
			bank                                          string
		)

		err := rows.Scan(&id, &companyName, &companyTIN, &date, &services, &totalValue, &operatorName, &documentType, &document, &documentIssuanceCity, &cellphone, &bankAccountNumber, &bankAccountType, &bank, &userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("DATABASE_ERROR", "Ocurrió un error al leer los datos de la base de datos"))
			return
		}

		servicesDeserialized, err := unmarshalServices(services)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("SERVICES_DESERIALIZATION_ERROR", "Error al deserializar servicios"))
			c.Abort()
			return
		}

		invoice := models.Invoice{
			ID:         id,
			Company:    models.Company{Name: companyName, TIN: companyTIN},
			Date:       date,
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
		c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponseInit("NOT_FOUND", "La página solicitada no existe"))
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al contar las facturas"))
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

func CreateInvoice(c *gin.Context) {
	// Obtener el rol y el ID de usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID

	// Comprobar si el usuario tiene el rol necesario para acceder a la ruta
	if !verifyRole(role, []string{common.ADMIN, common.USER}) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página."))
		return
	}

	var invoice models.Invoice
	if err := c.BindJSON(&invoice); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_DATA", "Datos inválidos. Verifica y vuelve a intentarlo."))
		return
	}

	servicesJSON, err := json.Marshal(invoice.Services)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("SERVICES_MARSHAL_ERROR", "Error al codificar los servicios en formato JSON."))
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al procesar las facturas"))
		c.Abort()
		return
	}

	// Actualiza el objeto de factura con el ID de usuario correcto
	invoice.UserID = userID

	c.JSON(http.StatusCreated, gin.H{"message": "Factura creada correctamente", "invoice": invoice})
}

// getUserIDFromInvoice consulta la base de datos para obtener el ID del usuario asociado a la factura especificada
func getUserIDFromInvoice(invoiceID string) (int64, error) {
	var invoiceUserID int64
	db := data.GetInstance()
	err := db.QueryRow("SELECT usuario_id FROM facturas WHERE id = $1", invoiceID).Scan(&invoiceUserID)
	if err != nil {
		return 0, err
	}
	return invoiceUserID, nil
}

func UpdateInvoice(c *gin.Context) {
	// Get the role and user ID from the JWT token
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID
	invoiceID := c.Param("id")

	// Check if the user has the necessary role to access the route
	if !verifyRole(role, []string{common.ADMIN, common.USER}) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página."))
		c.Abort()
		return
	}

	// Add a condition to allow common.ADMIN role to update any invoice
	if role != common.ADMIN {
		// Check if the user is trying to update their own invoice
		invoiceUserID, err := getUserIDFromInvoice(invoiceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al obtener el ID del usuario de la factura."))
			c.Abort()
			return
		}
		if userID != invoiceUserID {
			c.JSON(http.StatusForbidden, models.ErrorResponseInit("NO_PERMISSION", "Solo puedes actualizar tus propias facturas."))
			c.Abort()
			return
		}
	}

	var invoice models.Invoice
	err := c.BindJSON(&invoice)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_DATA", "Datos inválidos. Verifica y vuelve a intentarlo."))
		c.Abort()
		return
	}

	// Validate input data
	if invoice.Company.Name == "" || invoice.Company.TIN == "" || invoice.Date.IsZero() || len(invoice.Services) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("MISSING_FIELDS", "Faltan campos requeridos."))
		c.Abort()
		return
	}

	servicesJSON, err := json.Marshal(invoice.Services)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("SERVICES_MARSHAL_ERROR", "Error al codificar los servicios en formato JSON."))
		c.Abort()
		return
	}

	// Check if the user exists
	var userExists bool
	db := data.GetInstance()
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al verificar si el usuario existe."))
		c.Abort()
		return
	}
	if !userExists {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("USER_NOT_FOUND", "El usuario especificado no existe."))
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("DB_ERROR", "Error al actualizar la factura en la base de datos."))
		c.Abort()
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura actualizada correctamente"})
	} else {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit("INVOICE_NOT_FOUND", "No se encontró la factura con el ID especificado"))
	}
}

func DeleteInvoice(c *gin.Context) {
	// Get the role and user ID from the JWT token
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID

	// Check if the user has the necessary role to access the route
	allowedRoles := []string{common.ADMIN, common.USER}
	if !verifyRole(role, allowedRoles) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página."))
		c.Abort()
		return
	}

	// Validate the value of the id parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_ID", "El valor del parámetro id no es un número válido"))
		c.Abort()
		return
	}
	if id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_ID", "El valor del parámetro id debe ser un número entero positivo"))
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("SQL_ERROR", "Error al ejecutar la consulta SQL"))
		c.Abort()
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura eliminada correctamente"})
	} else {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit("INVOICE_NOT_FOUND", "No se encontró la factura con el ID especificado o no tienes permiso para eliminarla"))
		c.Abort()
	}
}

func getInvoice(c *gin.Context) (models.Invoice, error) {
	// Get the invoice ID from the URL parameter
	id := c.Param("id")

	// Query the database to get the invoice information
	query := `SELECT * FROM facturas WHERE id = $1`
	db := data.GetInstance()
	row := db.QueryRow(query, id)

	// Decode the row into an Invoice struct
	var invoice models.Invoice
	var servicesJSON []byte
	err := row.Scan(&invoice.ID, &invoice.Company.Name, &invoice.Company.TIN, &invoice.Date, &servicesJSON, &invoice.TotalValue, &invoice.Operator.Name, &invoice.Operator.DocumentType, &invoice.Operator.Document, &invoice.Operator.DocumentIssuanceCity, &invoice.Operator.Cellphone, &invoice.Operator.BankAccountNumber, &invoice.Operator.BankAccountType, &invoice.Operator.Bank, &invoice.UserID)

	if err != nil {
		if err == sql.ErrNoRows {
			//lint:ignore ST1005 Reason for ignoring warning
			return invoice, fmt.Errorf("No se encontró la factura con el ID especificado.")

		} else {
			return invoice, err
		}
	}

	// Decode the JSON data of the services into a slice of Service structs
	err = json.Unmarshal(servicesJSON, &invoice.Services)
	if err != nil {
		return invoice, err
	}

	return invoice, nil
}

func GeneratePDF(c *gin.Context) {
	// Get the invoice ID from the URL parameter
	id := c.Param("id")

	// Get the user role from the JWT token
	claims := c.MustGet("claims").(*models.Claims)
	role := claims.Role
	userID := claims.UserID

	// Check if the user has the necessary role to access the route
	allowedRoles := []string{common.ADMIN, common.USER}
	if !verifyRole(role, allowedRoles) {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página."))
		c.Abort()
		return
	}

	// Get the invoice
	invoice, err := getInvoice(c)
	if err != nil {
		if strings.Contains(err.Error(), "ID especificado.") {
			c.JSON(http.StatusNotFound, models.ErrorResponseInit("INVOICE_NOT_FOUND", err.Error()))
			return
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("INVOICE_NOT_FOUND", err.Error()))
			return
		}
	}

	// Check if the user is the owner of the invoice or has the ADMIN role
	if invoice.UserID != userID && role != common.ADMIN {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para generar el archivo PDF de esta factura."))
		c.Abort()
		return
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("DejaVuSans", "", "./font/DejaVuSans.ttf")
	pdf.AddPage()

	// Add invoice information
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "Cartagena   "+invoice.Date.Format("02-01-2006"))
	pdf.Ln(10)
	pdf.Cell(40, 10, invoice.Company.Name)
	pdf.Ln(10)
	pdf.Cell(40, 10, "Nit: "+invoice.Company.TIN)
	pdf.Ln(20)

	// Add client information
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "DEBE A:")
	pdf.Ln(10)
	pdf.Cell(40, 10, invoice.Operator.Name)
	pdf.Ln(10)
	pdf.Cell(40, 10, invoice.Operator.DocumentType+": "+invoice.Operator.Document+" Expedida en "+invoice.Operator.DocumentIssuanceCity)
	pdf.Ln(20)

	// Add total value
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "LA SUMA DE:")
	pdf.Ln(10)
	// Format the number with two decimal places
	formattedValue := strconv.FormatFloat(invoice.TotalValue, 'f', 2, 64)

	// Split the integer and decimal parts of the number
	parts := strings.Split(formattedValue, ".")

	// Add thousands separator to the integer part of the number
	for i := len(parts[0]) - 3; i > 0; i -= 3 {
		parts[0] = parts[0][:i] + "." + parts[0][i:]
	}

	// Join the integer and decimal parts of the number with a comma
	formattedValue = strings.Join(parts, ",")

	// Add peso symbol and "pesos" text
	text := fmt.Sprintf("$ %s pesos.", formattedValue)

	// Use formatted text in PDF cell
	pdf.Cell(40, 10, text)
	pdf.Ln(20)

	// Add concept
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "Por concepto de:")
	pdf.Ln(10)
	for _, service := range invoice.Services {
		pdf.Cell(80, 10, service.Description)
		pdf.Ln(10)
	}
	pdf.Ln(20)

	// Add signature and contact information
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "Cordialmente")
	pdf.Ln(20)
	pdf.Cell(40, 10, "_____________________________________________")
	pdf.Ln(20)
	pdf.Cell(40, 10, invoice.Operator.Name)
	pdf.Ln(10)
	pdf.Cell(40, 10, invoice.Operator.DocumentType+": "+invoice.Operator.Document)
	pdf.Ln(10)
	if invoice.Operator.Cellphone != "" {
		pdf.Cell(40, 10, "Cel: "+invoice.Operator.Cellphone)
	}
	if invoice.Operator.BankAccountNumber != "" {
		bankAccountType := ""
		if invoice.Operator.BankAccountType != "" {
			bankAccountType = " " + invoice.Operator.BankAccountType
		}
		bank := ""
		if invoice.Operator.Bank != "" {
			bank = " " + invoice.Operator.Bank
		}
		pdf.Cell(40, 10, fmt.Sprintf("N° Cuenta: %s%s%s", invoice.Operator.BankAccountNumber, bankAccountType, bank))
	}

	// Save the PDF to a temporary file
	tmpfile, err := os.CreateTemp("", fmt.Sprintf("factura-%s-*.pdf", id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(tmpfile.Name())

	err = pdf.Output(tmpfile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the content type and file name for download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="factura-%s.pdf"`, id))

	// Send the PDF file as a response
	c.File(tmpfile.Name())
}

func verifyRole(role string, allowedRoles []string) bool {
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}
