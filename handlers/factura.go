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

func unmarshalServicios(data []byte) ([]models.Servicio, error) {
	var servicios []models.Servicio
	err := json.Unmarshal(data, &servicios)
	if err != nil {
		return nil, err
	}
	return servicios, nil
}

func ListarFacturas(c *gin.Context, db *data.DB) {
	// Obtener el rol del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if rol != common.ADMIN && rol != common.USER {
		errorResponse := models.ErrorResponseInit("INSUFFICIENT_ROLE", "Lo siento, pero parece que no tienes los permisos necesarios para acceder a esta página. Por favor, verifica tus credenciales o contacta al ADMIN para obtener más información.")
		c.AbortWithStatusJSON(http.StatusForbidden, errorResponse)
		return
	}

	// Obtener y validar los parámetros de la consulta
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		errorResponse := models.ErrorResponseInit("INVALID_PAGE_PARAM", "El parámetro 'page' debe ser un número entero positivo")
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse)
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		errorResponse := models.ErrorResponseInit("INVALID_LIMIT_PARAM", "El parámetro 'limit' debe ser un número entero positivo")
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse)
		return
	}
	if limit > 100 {
		errorResponse := models.ErrorResponseInit("LIMIT_TOO_HIGH", "El parámetro 'limit' no puede ser mayor a 100")
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse)
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
			args = []interface{}{claims.UsuarioID, limit, offset}
		}
	}
	rows, err = db.Query(query, args...)
	if err != nil {
		errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al obtener las facturas")
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	defer rows.Close()

	// Procesar las filas y construir el arreglo de facturas
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
			errorResponse := models.ErrorResponseInit("DATABASE_ERROR", "Ocurrió un error al leer los datos de la base de datos")
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}

		serviciosDeserializados, err := unmarshalServicios(servicios)
		if err != nil {
			errorResponse := models.ErrorResponseInit("SERVICES_DESERIALIZATION_ERROR", "Error al deserializar servicios")
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			c.Abort()
			return
		}

		factura := models.Factura{
			ID:         id,
			Empresa:    models.Empresa{Nombre: nombreEmpresa, NIT: nitEmpresa},
			Fecha:      fecha,
			Servicios:  serviciosDeserializados,
			ValorTotal: valorTotal,
			Operador:   models.Operador{Nombre: nombreOperador, TipoDocumento: tipoDocumento, Documento: documento, CiudadExpedicionDocumento: ciudadExpedicion, Celular: celular, NumeroCuentaBancaria: numeroCuenta, TipoCuentaBancaria: tipoCuenta, Banco: banco},
			UsuarioID:  int64(usuarioID),
		}
		facturas = append(facturas, factura)
	}

	// Verificar si se encontraron facturas y contar el total de facturas
	switch len(facturas) {
	case 0:
		errorResponse := models.ErrorResponseInit("NOT_FOUND", "La página solicitada no existe")
		c.AbortWithStatusJSON(http.StatusNotFound, errorResponse)
		c.Abort()
	default:
		var totalFacturas int
		if filterField != "" && filterValue != "" {
			query = `SELECT COUNT(*) FROM facturas WHERE $1 = $2`
			args = []interface{}{filterField, filterValue}
		} else {
			if rol == common.ADMIN {
				query = `SELECT COUNT(*) FROM facturas`
				args = []interface{}{}
			} else {
				query = `SELECT COUNT(*) FROM facturas WHERE usuario_id = $1`
				args = []interface{}{claims.UsuarioID}
			}
		}
		err = db.QueryRow(query, args...).Scan(&totalFacturas)
		if err != nil {
			errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al contar las facturas")
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			c.Abort()
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

func CrearFactura(c *gin.Context, db *data.DB) {
	// Obtener el rol y el ID del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role
	idUsuario := claims.UsuarioID

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if !verificarRol(rol, []string{common.ADMIN, common.USER}) {
		errorResponse := models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página.")
		c.JSON(http.StatusForbidden, errorResponse)
		return
	}

	var factura models.Factura
	if err := c.BindJSON(&factura); err != nil {
		errorResponse := models.ErrorResponseInit("INVALID_DATA", "Datos inválidos. Verifica y vuelve a intentarlo.")
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	serviciosJSON, err := json.Marshal(factura.Servicios)
	if err != nil {
		errorResponse := models.ErrorResponseInit("SERVICES_MARSHAL_ERROR", "Error al codificar los servicios en formato JSON.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	query := `INSERT INTO facturas (nombre_empresa, nit_empresa, fecha, servicios, valor_total, nombre_operador, tipo_documento_operador, documento_operador, ciudad_expedicion_documento_operador, celular_operador, numero_cuenta_bancaria_operador, tipo_cuenta_bancaria_operador, banco_operador, usuario_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,$14) RETURNING id`
	err = db.QueryRow(query,
		factura.Empresa.Nombre,
		factura.Empresa.NIT,
		factura.Fecha,
		serviciosJSON,
		factura.ValorTotal,
		factura.Operador.Nombre,
		factura.Operador.TipoDocumento,
		factura.Operador.Documento,
		factura.Operador.CiudadExpedicionDocumento,
		factura.Operador.Celular,
		factura.Operador.NumeroCuentaBancaria,
		factura.Operador.TipoCuentaBancaria,
		factura.Operador.Banco,
		idUsuario).Scan(&factura.ID)

	if err != nil {
		errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al procesar las facturas")
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	// Actualizar el objeto factura con el idUsuario correcto
	factura.UsuarioID = idUsuario

	c.JSON(http.StatusCreated, gin.H{"message": "Factura creada correctamente", "factura": factura})
}

// obtenerUsuarioIDFactura consulta la base de datos para obtener el ID del usuario asociado a la factura especificada
func obtenerUsuarioIDFactura(db *data.DB, idFactura string) (int64, error) {
	var idUsuarioFactura int64
	err := db.QueryRow("SELECT usuario_id FROM facturas WHERE id = $1", idFactura).Scan(&idUsuarioFactura)
	if err != nil {
		return 0, err
	}
	return idUsuarioFactura, nil
}

func ActualizarFactura(c *gin.Context, db *data.DB) {
	// Obtener el rol y el ID del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role
	idUsuario := claims.UsuarioID
	idFactura := c.Param("id")

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	if !verificarRol(rol, []string{common.ADMIN, common.USER}) {
		errorResponse := models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página.")
		c.JSON(http.StatusForbidden, errorResponse)
		c.Abort()
		return
	}

	// Agregar una condición para permitir que el rol de common.ADMIN actualice cualquier factura
	if rol != common.ADMIN {
		// Verificar si el usuario está intentando actualizar su propia factura
		idUsuarioFactura, err := obtenerUsuarioIDFactura(db, idFactura)
		if err != nil {
			errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al obtener el ID del usuario de la factura.")
			c.JSON(http.StatusInternalServerError, errorResponse)
			c.Abort()
			return
		}
		if idUsuario != idUsuarioFactura {
			errorResponse := models.ErrorResponseInit("NO_PERMISSION", "Solo puedes actualizar tus propias facturas.")
			c.JSON(http.StatusForbidden, errorResponse)
			c.Abort()
			return
		}
	}

	var factura models.Factura
	err := c.BindJSON(&factura)
	if err != nil {
		errorResponse := models.ErrorResponseInit("INVALID_DATA", "Datos inválidos. Verifica y vuelve a intentarlo.")
		c.JSON(http.StatusBadRequest, errorResponse)
		c.Abort()
		return
	}

	// Validar los datos de entrada
	if factura.Empresa.Nombre == "" || factura.Empresa.NIT == "" || factura.Fecha.IsZero() || len(factura.Servicios) == 0 {
		errorResponse := models.ErrorResponseInit("MISSING_FIELDS", "Faltan campos requeridos.")
		c.JSON(http.StatusBadRequest, errorResponse)
		c.Abort()
		return
	}

	serviciosJSON, err := json.Marshal(factura.Servicios)
	if err != nil {
		errorResponse := models.ErrorResponseInit("SERVICIOS_MARSHAL_ERROR", "Error al codificar los servicios en formato JSON.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	// Verificar si el usuario existe
	var existeUsuario bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE id = $1)", idUsuario).Scan(&existeUsuario)
	if err != nil {
		errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al verificar si el usuario existe.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}
	if !existeUsuario {
		errorResponse := models.ErrorResponseInit("USER_NOT_FOUND", "El usuario especificado no existe.")
		c.JSON(http.StatusBadRequest, errorResponse)
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
	result, err := db.Exec(query, factura.Empresa.Nombre, factura.Empresa.NIT, factura.Fecha, serviciosJSON, factura.ValorTotal, factura.Operador.Nombre, factura.Operador.TipoDocumento, factura.Operador.Documento, factura.Operador.CiudadExpedicionDocumento, factura.Operador.Celular, factura.Operador.NumeroCuentaBancaria, factura.Operador.TipoCuentaBancaria, factura.Operador.Banco, idFactura)
	if err != nil {
		errorResponse := models.ErrorResponseInit("DB_ERROR", "Error al actualizar la factura en la base de datos.")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura actualizada correctamente"})
	} else {
		errorResponse := models.ErrorResponseInit("INVOICE_NOT_FOUND", "No se encontró la factura con el ID especificado")
		c.JSON(http.StatusNotFound, errorResponse)
	}
}

func EliminarFactura(c *gin.Context, db *data.DB) {
	// Obtener el rol y usuario_id del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role
	idUsuario := claims.UsuarioID

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	rolesPermitidos := []string{common.ADMIN, common.USER}
	if !verificarRol(rol, rolesPermitidos) {
		errorResponse := models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página.")
		c.JSON(http.StatusForbidden, errorResponse)
		c.Abort()
		return
	}

	// Validar el valor del parámetro id
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse := models.ErrorResponseInit("INVALID_ID", "El valor del parámetro id no es un número válido")
		c.JSON(http.StatusBadRequest, errorResponse)
		c.Abort()
		return
	}
	if id <= 0 {
		errorResponse := models.ErrorResponseInit("INVALID_ID", "El valor del parámetro id debe ser un número entero positivo")
		c.JSON(http.StatusBadRequest, errorResponse)
		c.Abort()
		return
	}

	var query string
	if rol == common.ADMIN {
		query = `DELETE FROM facturas WHERE id = $1`
	} else {
		query = `DELETE FROM facturas WHERE id = $1 AND usuario_id = $2`
	}

	var result sql.Result
	if rol == common.ADMIN {
		result, err = db.Exec(query, id)
	} else {
		result, err = db.Exec(query, id, idUsuario)
	}
	if err != nil {
		errorResponse := models.ErrorResponseInit("SQL_ERROR", "Error al ejecutar la consulta SQL")
		c.JSON(http.StatusInternalServerError, errorResponse)
		c.Abort()
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Factura eliminada correctamente"})
	} else {
		errorResponse := models.ErrorResponseInit("INVOICE_NOT_FOUND", "No se encontró la factura con el ID especificado o no tienes permiso para eliminarla")
		c.JSON(http.StatusNotFound, errorResponse)
		c.Abort()
	}
}

func obtenerFactura(c *gin.Context, db *data.DB) (models.Factura, error) {
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
			//lint:ignore ST1005 Razón para ignorar la advertencia
			return factura, fmt.Errorf("No se encontró la factura con el ID especificado.")

		} else {
			return factura, err
		}
	}

	// Decodificar los datos JSON de los servicios en una slice de estructuras Servicio
	err = json.Unmarshal(serviciosJSON, &factura.Servicios)
	if err != nil {
		return factura, err
	}

	return factura, nil
}

func GenerarPDF(c *gin.Context, db *data.DB) {
	// Obtener el ID de la factura a partir del parámetro de la URL
	id := c.Param("id")

	// Obtener el rol del usuario del token JWT
	claims := c.MustGet("claims").(*models.Claims)
	rol := claims.Role
	idUsuario := claims.UsuarioID

	// Verificar si el usuario tiene el rol necesario para acceder a la ruta
	rolesPermitidos := []string{common.ADMIN, common.USER}
	if !verificarRol(rol, rolesPermitidos) {
		errorResponse := models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para acceder a esta página.")
		c.JSON(http.StatusForbidden, errorResponse)
		c.Abort()
		return
	}

	// Obtener la factura
	factura, err := obtenerFactura(c, db)
	if err != nil {
		if strings.Contains(err.Error(), "ID especificado.") {
			c.JSON(http.StatusNotFound, models.ErrorResponseInit("INVOICE_NOT_FOUND", err.Error()))
			return
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("INVOICE_NOT_FOUND", err.Error()))
			return
		}
	}

	// Verificar si el usuario es el dueño de la factura o si tiene el rol de ADMIN
	if factura.UsuarioID != idUsuario && rol != common.ADMIN {
		errorResponse := models.ErrorResponseInit("NO_PERMISSION", "No tienes permiso para generar el archivo PDF de esta factura.")
		c.JSON(http.StatusForbidden, errorResponse)
		c.Abort()
		return
	}

	// Crear un nuevo documento PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("DejaVuSans", "", "./font/DejaVuSans.ttf")
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

func verificarRol(rol string, rolesPermitidos []string) bool {
	for _, rolPermitido := range rolesPermitidos {
		if rol == rolPermitido {
			return true
		}
	}
	return false
}
