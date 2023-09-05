package helpers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
)

func CheckRoleExists(db *data.PostgresAdapter, roleName string) error {
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM roles WHERE name=$1`)
	if err != nil {
		return models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta.")
	}
	defer stmt.Close()
	row := stmt.QueryRow(roleName)
	var count int
	if err = row.Scan(&count); err != nil || count == 0 {
		return models.ErrorResponseInit(common.ErrRoleNotFound, "No se encontró el rol especificado.")
	}
	return nil
}
