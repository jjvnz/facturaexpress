package helpers

import "facturaexpress/data"

// getUserIDFromInvoice consulta la base de datos para obtener el ID del usuario asociado a la factura especificada
func GetUserIDFromInvoice(invoiceID string) (int64, error) {
	var invoiceUserID int64
	db := data.GetInstance()
	err := db.QueryRow("SELECT usuario_id FROM facturas WHERE id = $1", invoiceID).Scan(&invoiceUserID)
	if err != nil {
		return 0, err
	}
	return invoiceUserID, nil
}
