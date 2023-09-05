package handlers

import (
	"database/sql"
	"encoding/json"
	"facturaexpress/data"
	"facturaexpress/models"
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetInvoice(c *gin.Context) (models.Invoice, error) {
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
