package models

import "time"

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
