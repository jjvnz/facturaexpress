package models

import "time"

type Factura struct {
	ID                                int        `json:"id"`
	NombreEmpresa                     string     `json:"nombre_empresa"`
	NITEmpresa                        string     `json:"nit_empresa"`
	Fecha                             time.Time  `json:"fecha"`
	Servicios                         []Servicio `json:"servicios"`
	ValorTotal                        float64    `json:"valor_total"`
	NombreOperador                    string     `json:"nombre_operador"`
	TipoDocumentoOperador             string     `json:"tipo_documento_operador"`
	DocumentoOperador                 string     `json:"documento_operador"`
	CiudadExpedicionDocumentoOperador string     `json:"ciudad_expedicion_documento_operador"`
	CelularOperador                   string     `json:"celular_operador"`
	NumeroCuentaBancariaOperador      string     `json:"numero_cuenta_bancaria_operador"`
	TipoCuentaBancariaOperador        string     `json:"tipo_cuenta_bancaria_operador"`
	BancoOperador                     string     `json:"banco_operador"`
}

type Servicio struct {
	Descripcion string  `json:"descripcion"`
	Valor       float64 `json:"valor"`
}
