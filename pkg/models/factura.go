package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Factura struct {
	ID         int        `json:"id"`
	Empresa    Empresa    `json:"empresa"`
	Fecha      time.Time  `json:"fecha"`
	Servicios  []Servicio `json:"servicios"`
	ValorTotal float64    `json:"valor_total"`
	Operador   Operador   `json:"operador"`
	UsuarioID  int64      `json:"usuario_id"`
}

type Empresa struct {
	Nombre string `json:"nombre"`
	NIT    string `json:"nit"`
}

type Operador struct {
	Nombre                    string `json:"nombre"`
	TipoDocumento             string `json:"tipo_documento"`
	Documento                 string `json:"documento"`
	CiudadExpedicionDocumento string `json:"ciudad_expedicion_documento"`
	Celular                   string `json:"celular"`
	NumeroCuentaBancaria      string `json:"numero_cuenta_bancaria"`
	TipoCuentaBancaria        string `json:"tipo_cuenta_bancaria"`
	Banco                     string `json:"banco"`
}

type Servicio struct {
	Descripcion string  `json:"descripcion"`
	Valor       float64 `json:"valor"`
}

func unmarshalServicios(data []byte) []Servicio {
	var servicios []Servicio
	err := json.Unmarshal(data, &servicios)
	if err != nil {
		fmt.Println("Error al decodificar los datos JSON!", err)
	}
	return servicios
}

func (f *Factura) Scan(row *sql.Rows) error {
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
	err := row.Scan(&id, &nombreEmpresa, &nitEmpresa, &fecha, &servicios, &valorTotal, &nombreOperador, &tipoDocumento, &documento, &ciudadExpedicion, &celular, &numeroCuenta, &tipoCuenta, &banco, &usuarioID)
	if err != nil {
		return err
	}
	f.ID = id
	f.Empresa = Empresa{Nombre: nombreEmpresa, NIT: nitEmpresa}
	f.Fecha = fecha
	f.Servicios = unmarshalServicios(servicios)
	f.ValorTotal = valorTotal
	f.Operador = Operador{Nombre: nombreOperador,
		TipoDocumento:             tipoDocumento,
		Documento:                 documento,
		CiudadExpedicionDocumento: ciudadExpedicion,
		Celular:                   celular,
		NumeroCuentaBancaria:      numeroCuenta,
		TipoCuentaBancaria:        tipoCuenta,
		Banco:                     banco}
	f.UsuarioID = int64(usuarioID)
	return nil
}
