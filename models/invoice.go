package models

import (
	"time"
)

type Invoice struct {
	ID         int       `json:"id"`
	Company    Company   `json:"empresa"`
	Date       time.Time `json:"fecha"`
	Services   []Service `json:"servicios"`
	TotalValue float64   `json:"valor_total"`
	Operator   Operator  `json:"operador"`
	UserID     int64     `json:"usuario_id"`
}

type Company struct {
	Name string `json:"nombre"`
	TIN  string `json:"nit"`
}

type Operator struct {
	Name                 string `json:"nombre"`
	DocumentType         string `json:"tipo_documento"`
	Document             string `json:"documento"`
	DocumentIssuanceCity string `json:"ciudad_expedicion_documento"`
	Cellphone            string `json:"celular"`
	BankAccountNumber    string `json:"numero_cuenta_bancaria"`
	BankAccountType      string `json:"tipo_cuenta_bancaria"`
	Bank                 string `json:"banco"`
}

type Service struct {
	Description string  `json:"descripcion"`
	Value       float64 `json:"valor"`
}
