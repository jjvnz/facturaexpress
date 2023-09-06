package handlers

import (
	"facturaexpress/common"
	"facturaexpress/helpers"
	"facturaexpress/models"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(c *gin.Context) {
	// Get the invoice ID from the URL parameter
	id := c.Param("id")

	// Get the user role from the JWT token
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

	// Get the invoice
	invoice, err := GetInvoice(c)
	if err != nil {
		if strings.Contains(err.Error(), "ID especificado.") {
			c.JSON(http.StatusNotFound, models.ErrorResponseInit(common.ErrInvoiceNotFound, err.Error()))
			return
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit(common.ErrInvoiceNotFound, err.Error()))
			return
		}
	}

	// Check if the user is the owner of the invoice or has the ADMIN role
	if invoice.UserID != userID && role != common.ADMIN {
		c.JSON(http.StatusForbidden, models.ErrorResponseInit(common.ErrNoPermission, "No tienes permiso para generar el archivo PDF de esta factura."))
		c.Abort()
		return
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("DejaVuSans", "", "./font/DejaVuSans.ttf")
	pdf.AddPage()

	// Add invoice information
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "Cartagena "+helpers.FormatDateInSpanish(invoice.Date))
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
