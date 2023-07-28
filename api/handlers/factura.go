package handlers

import (
	"database/sql"
	"encoding/json"
	"facturaexpress/pkg/models"
	"facturaexpress/pkg/storage"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func ListarFacturas(c *gin.Context, db *storage.DB) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'page' debe ser un número entero positivo"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'limit' debe ser un número entero positivo"})
		return
	}

	if limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'limit' no puede ser mayor a 100"})
		return
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	filterField := c.Query("filter_field")
	filterValue := c.Query("filter_value")

	var rows *sql.Rows
	if filterField != "" && filterValue != "" {
		query := `SELECT * FROM facturas WHERE $1 = $2 ORDER BY id ASC LIMIT $3 OFFSET $4`
		rows, err = db.Query(query, filterField, filterValue, limit, offset)
	} else {
		query := `SELECT * FROM facturas ORDER BY id ASC LIMIT $1 OFFSET $2`
		rows, err = db.Query(query, limit, offset)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Loop through the rows and decode each row into a Factura struct

	var facturas []models.Factura
	for rows.Next() {
		var factura models.Factura
		var serviciosJSON []byte
		err := rows.Scan(&factura.ID, &factura.Empresa.Nombre, &factura.Empresa.NIT, &factura.Fecha, &serviciosJSON, &factura.ValorTotal, &factura.Operador.Nombre, &factura.Operador.TipoDocumento, &factura.Operador.Documento, &factura.Operador.CiudadExpedicionDocumento, &factura.Operador.Celular, &factura.Operador.NumeroCuentaBancaria, &factura.Operador.TipoCuentaBancaria, &factura.Operador.Banco)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = json.Unmarshal(serviciosJSON, &factura.Servicios)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		facturas = append(facturas, factura)
	}

	switch len(facturas) {
	case 0:
		c.JSON(http.StatusNotFound, gin.H{"error": "La página solicitada no existe"})
	default:
		// Count the total number of facturas in the database
		var totalFacturas int
		err = db.QueryRow(`SELECT COUNT(*) FROM facturas`).Scan(&totalFacturas)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Calculate the total number of pages
		totalPages := int(math.Ceil(float64(totalFacturas) / float64(limit)))

		c.JSON(http.StatusOK, gin.H{
			"facturas":    facturas,
			"total_pages": totalPages,
			"page":        page,
		})
	}
}

func CrearFactura(c *gin.Context, db *storage.DB) {
	// Decodifica el cuerpo de la solicitud para obtener los datos de la factura
	var factura models.Factura
	err := c.BindJSON(&factura)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convierte la slice de estructuras a una cadena JSON
	serviciosJSON, err := json.Marshal(factura.Servicios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Guarda la factura en la base de datos y recupera el ID generado
	query := `INSERT INTO facturas (nombre_empresa, nit_empresa, fecha, servicios, valor_total, nombre_operador, tipo_documento_operador, documento_operador, ciudad_expedicion_documento_operador, celular_operador, numero_cuenta_bancaria_operador, tipo_cuenta_bancaria_operador, banco_operador) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`
	err = db.QueryRow(query, factura.Empresa.Nombre, factura.Empresa.NIT, factura.Fecha, serviciosJSON, factura.ValorTotal, factura.Operador.Nombre, factura.Operador.TipoDocumento, factura.Operador.Documento, factura.Operador.CiudadExpedicionDocumento, factura.Operador.Celular, factura.Operador.NumeroCuentaBancaria, factura.Operador.TipoCuentaBancaria, factura.Operador.Banco).Scan(&factura.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Factura creada correctamente", "factura": factura})
}

func ActualizarFactura(c *gin.Context, db *storage.DB) {
	// Get the ID of the factura to update from the URL parameter
	id := c.Param("id")

	// Decode the request body to get the updated factura data
	var factura models.Factura
	err := c.BindJSON(&factura)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert the slice of Servicio structs to a JSON string
	serviciosJSON, err := json.Marshal(factura.Servicios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update the factura in the database
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
	result, err := db.Exec(query, factura.Empresa.Nombre, factura.Empresa.NIT, factura.Fecha, serviciosJSON, factura.ValorTotal, factura.Operador.Nombre, factura.Operador.TipoDocumento, factura.Operador.Documento, factura.Operador.CiudadExpedicionDocumento, factura.Operador.Celular, factura.Operador.NumeroCuentaBancaria, factura.Operador.TipoCuentaBancaria, factura.Operador.Banco, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura actualizada correctamente"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "No se encontró la factura con el ID especificado"})
	}
}

func EliminarFactura(c *gin.Context, db *storage.DB) {
	// Get the ID of the factura to delete from the URL parameter
	id := c.Param("id")

	// Delete the factura from the database
	query := `DELETE FROM facturas WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		// Return a success message if the delete was successful
		c.JSON(http.StatusOK, gin.H{"message": "Factura eliminada correctamente"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "No se encontró la factura con el ID especificado"})
	}
}

func GenerarPDF(c *gin.Context, db *storage.DB) {
	// Obtener el ID de la factura a partir del parámetro de la URL
	id := c.Param("id")

	// Consultar la base de datos para obtener la información de la factura
	query := `SELECT * FROM facturas WHERE id = $1`
	row := db.QueryRow(query, id)

	// Decodificar la fila en una estructura Factura
	var factura models.Factura
	var serviciosJSON []byte
	err := row.Scan(&factura.ID, &factura.Empresa.Nombre, &factura.Empresa.NIT, &factura.Fecha, &serviciosJSON, &factura.ValorTotal, &factura.Operador.Nombre, &factura.Operador.TipoDocumento, &factura.Operador.Documento, &factura.Operador.CiudadExpedicionDocumento, &factura.Operador.Celular, &factura.Operador.NumeroCuentaBancaria, &factura.Operador.TipoCuentaBancaria, &factura.Operador.Banco)

	if err != nil {
		if err == sql.ErrNoRows {
			// Manejar el caso en que no hay filas para escanear
			c.JSON(http.StatusNotFound, gin.H{"error": "Factura no encontrada"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Decodificar los datos JSON de los servicios en una slice de estructuras Servicio
	err = json.Unmarshal(serviciosJSON, &factura.Servicios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crear un nuevo documento PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Agregar información de la factura
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Cartagena   "+factura.Fecha.Format("02-01-2006"))
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Empresa.Nombre)
	pdf.Ln(10)
	pdf.Cell(40, 10, "Nit: "+factura.Empresa.NIT)
	pdf.Ln(20)

	// Agregar información del cliente
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "DEBE A:")
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Operador.Nombre)
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Operador.TipoDocumento+": "+factura.Operador.Documento+" Expedida en "+factura.Operador.CiudadExpedicionDocumento)
	pdf.Ln(20)

	// Agregar valor total
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "LA SUMA DE:")
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("$ %.2f pesos.", factura.ValorTotal))
	pdf.Ln(20)

	// Agregar concepto
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Por concepto de:")
	pdf.Ln(10)
	for _, servicio := range factura.Servicios {
		pdf.Cell(80, 10, servicio.Descripcion)
		pdf.Ln(10)
	}
	pdf.Ln(20)

	// Agregar firma y datos de contacto
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Cordialmente")
	pdf.Ln(20)
	pdf.Cell(40, 10, "_____________________________________________")
	pdf.Ln(20)
	pdf.Cell(40, 10, factura.Operador.Nombre)
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Operador.TipoDocumento+": "+factura.Operador.Documento)
	pdf.Ln(10)
	if factura.Operador.Celular != "" {
		pdf.Cell(40, 10, "Cel: "+factura.Operador.Celular)
	}
	if factura.Operador.NumeroCuentaBancaria != "" {
		tipoCuenta := ""
		if factura.Operador.TipoCuentaBancaria != "" {
			tipoCuenta = " " + factura.Operador.TipoCuentaBancaria
		}
		banco := ""
		if factura.Operador.Banco != "" {
			banco = " " + factura.Operador.Banco
		}
		pdf.Cell(40, 10, fmt.Sprintf("N° Cuenta: %s%s%s", factura.Operador.NumeroCuentaBancaria, tipoCuenta, banco))
	}

	// Guardar el PDF en un archivo temporal
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

	// Establecer el tipo de contenido y el nombre del archivo para la descarga
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="factura-%s.pdf"`, id))

	// Enviar el archivo PDF como respuesta
	c.File(tmpfile.Name())
}
