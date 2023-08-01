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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func unmarshalServicios(data []byte) []models.Servicio {
	var servicios []models.Servicio
	err := json.Unmarshal(data, &servicios)
	if err != nil {
		fmt.Println("error al decodificar los datos JSON!", err)
	}
	return servicios
}

func ListarFacturas(c *gin.Context, db *storage.DB) {
	// Obtener el rol del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if rol != "administrador" && rol != "usuario" {
		c.JSON(http.StatusForbidden, gin.H{"error": "El usuario no tiene el rol necesario para acceder a esta ruta"})
		return
	}
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
		// Verificar si el usuario tiene el rol de "administrador"
		if rol == "administrador" {
			// Si el usuario es un administrador, listar todas las facturas
			query := `SELECT * FROM facturas ORDER BY id ASC LIMIT $1 OFFSET $2`
			rows, err = db.Query(query, limit, offset)
		} else {
			// Si el usuario no es un administrador, listar solo sus propias facturas
			query := `SELECT * FROM facturas WHERE usuario_id = $1 ORDER BY id ASC LIMIT $2 OFFSET $3`
			rows, err = db.Query(query, claims.UsuarioID, limit, offset)
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var facturas []models.Factura
	for rows.Next() {
		var (
			id, usuarioID                     int
			nombreEmpresa, nitEmpresa         string
			fecha                             time.Time
			servicios                         []byte
			valorTotal                        float64
			nombreOperador, tipoDocumento     string
			documento, ciudadExpedicion       string
			celular, numeroCuenta, tipoCuenta string
			banco                             string
		)
		err := rows.Scan(&id, &nombreEmpresa, &nitEmpresa, &fecha, &servicios, &valorTotal, &nombreOperador, &tipoDocumento, &documento, &ciudadExpedicion, &celular, &numeroCuenta, &tipoCuenta, &banco, &usuarioID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		factura := models.Factura{
			ID:         id,
			Empresa:    models.Empresa{Nombre: nombreEmpresa, NIT: nitEmpresa},
			Fecha:      fecha,
			Servicios:  unmarshalServicios(servicios),
			ValorTotal: valorTotal,
			Operador: models.Operador{Nombre: nombreOperador, TipoDocumento: tipoDocumento,
				Documento: documento, CiudadExpedicionDocumento: ciudadExpedicion,
				Celular:              celular,
				NumeroCuentaBancaria: numeroCuenta,
				TipoCuentaBancaria:   tipoCuenta, Banco: banco},
			UsuarioID: int64(usuarioID),
		}
		facturas = append(facturas, factura)
	}
	switch len(facturas) {
	case 0:
		c.JSON(http.StatusNotFound, gin.H{"error": "La página solicitada no existe"})
	default:
		var totalFacturas int
		// Verificar si el usuario tiene el rol de "administrador"
		if rol == "administrador" {
			// Si el usuario es un administrador contar todas las facturas
			err = db.QueryRow(`SELECT COUNT(*) FROM facturas`).Scan(&totalFacturas)
		} else {
			// Si el usuario no es un administrador contar solo sus propias facturas
			query := `SELECT COUNT(*) FROM facturas WHERE usuario_id = $1`
			err = db.QueryRow(query, claims.UsuarioID).Scan(&totalFacturas)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		totalPages := int(math.Ceil(float64(totalFacturas) / float64(limit)))
		c.JSON(http.StatusOK, gin.H{
			"facturas":    facturas,
			"total_pages": totalPages,
			"page":        page,
		})
	}
}

func CrearFactura(c *gin.Context, db *storage.DB) {
	// Obtener el rol del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if rol != "administrador" && rol != "usuario" {
		c.JSON(http.StatusForbidden, gin.H{"error": "El usuario no tiene el rol necesario para acceder a esta ruta"})
		return
	}
	var factura models.Factura
	err := c.BindJSON(&factura)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviciosJSON, err := json.Marshal(factura.Servicios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	query := `INSERT INTO facturas (nombre_empresa, nit_empresa, fecha, servicios, valor_total, nombre_operador, tipo_documento_operador, documento_operador, ciudad_expedicion_documento_operador, celular_operador, numero_cuenta_bancaria_operador, tipo_cuenta_bancaria_operador, banco_operador, usuario_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,$14) RETURNING id`
	err = db.QueryRow(query, factura.Empresa.Nombre, factura.Empresa.NIT, factura.Fecha, serviciosJSON, factura.ValorTotal, factura.Operador.Nombre, factura.Operador.TipoDocumento, factura.Operador.Documento, factura.Operador.CiudadExpedicionDocumento, factura.Operador.Celular, factura.Operador.NumeroCuentaBancaria, factura.Operador.TipoCuentaBancaria, factura.Operador.Banco, factura.UsuarioID).Scan(&factura.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Factura creada correctamente", "factura": factura})
}

func ActualizarFactura(c *gin.Context, db *storage.DB) {
	id := c.Param("id")

	var factura models.Factura
	err := c.BindJSON(&factura)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviciosJSON, err := json.Marshal(factura.Servicios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
    banco_operador = $13, usuario_id=$14 WHERE id = $15`
	result, err := db.Exec(query, factura.Empresa.Nombre, factura.Empresa.NIT, factura.Fecha, serviciosJSON, factura.ValorTotal, factura.Operador.Nombre, factura.Operador.TipoDocumento, factura.Operador.Documento, factura.Operador.CiudadExpedicionDocumento, factura.Operador.Celular, factura.Operador.NumeroCuentaBancaria, factura.Operador.TipoCuentaBancaria, factura.Operador.Banco, factura.UsuarioID, id)
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
	id := c.Param("id")

	query := `DELETE FROM facturas WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
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
	err := row.Scan(&factura.ID, &factura.Empresa.Nombre, &factura.Empresa.NIT, &factura.Fecha, &serviciosJSON, &factura.ValorTotal, &factura.Operador.Nombre, &factura.Operador.TipoDocumento, &factura.Operador.Documento, &factura.Operador.CiudadExpedicionDocumento, &factura.Operador.Celular, &factura.Operador.NumeroCuentaBancaria, &factura.Operador.TipoCuentaBancaria, &factura.Operador.Banco, &factura.UsuarioID)

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
	pdf.AddUTF8Font("DejaVuSans", "", "../../font/DejaVuSans.ttf")
	pdf.AddPage()

	// Agregar información de la factura
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "Cartagena   "+factura.Fecha.Format("02-01-2006"))
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Empresa.Nombre)
	pdf.Ln(10)
	pdf.Cell(40, 10, "Nit: "+factura.Empresa.NIT)
	pdf.Ln(20)

	// Agregar información del cliente
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "DEBE A:")
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Operador.Nombre)
	pdf.Ln(10)
	pdf.Cell(40, 10, factura.Operador.TipoDocumento+": "+factura.Operador.Documento+" Expedida en "+factura.Operador.CiudadExpedicionDocumento)
	pdf.Ln(20)

	// Agregar valor total
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "LA SUMA DE:")
	pdf.Ln(10)
	// Formatear el número con dos decimales
	valorFormateado := strconv.FormatFloat(factura.ValorTotal, 'f', 2, 64)

	// Separar la parte entera y la parte decimal del número
	partes := strings.Split(valorFormateado, ".")

	// Agregar el separador de miles a la parte entera del número
	for i := len(partes[0]) - 3; i > 0; i -= 3 {
		partes[0] = partes[0][:i] + "." + partes[0][i:]
	}

	// Unir la parte entera y la parte decimal del número con una coma
	valorFormateado = strings.Join(partes, ",")

	// Agregar el símbolo de peso y el texto "pesos"
	texto := fmt.Sprintf("$ %s pesos.", valorFormateado)

	// Utilizar el texto formateado en la celda del PDF
	pdf.Cell(40, 10, texto)
	pdf.Ln(20)

	// Agregar concepto
	pdf.SetFont("DejaVuSans", "", 12)
	pdf.Cell(40, 10, "Por concepto de:")
	pdf.Ln(10)
	for _, servicio := range factura.Servicios {
		pdf.Cell(80, 10, servicio.Descripcion)
		pdf.Ln(10)
	}
	pdf.Ln(20)

	// Agregar firma y datos de contacto
	pdf.SetFont("DejaVuSans", "", 12)
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
