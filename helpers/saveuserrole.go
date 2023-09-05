package helpers

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
)

func SaveUserRole(db *data.PostgresAdapter, userID int64) error {
	stmt, err := db.Prepare(`SELECT id FROM roles WHERE name=$1`)
	if err != nil {
		return models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta.")
	}
	defer stmt.Close()
	row := stmt.QueryRow(common.USER)
	var roleID int64
	if err = row.Scan(&roleID); err != nil {
		if err == sql.ErrNoRows {
			return models.ErrorResponseInit(common.ErrRoleNotFound, "No se encontró el rol especificado.")
		}
		return models.ErrorResponseInit(common.ErrRoleIDRetrievalFailed, "Error al obtener el ID del rol.")
	}

	stmt, err = db.Prepare(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`)
	if err != nil {
		return models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta.")
	}
	defer stmt.Close()
	if _, err = stmt.Exec(userID, roleID); err != nil {
		return models.ErrorResponseInit(common.ErrDatabaseSaveFailed, "Error al guardar en la base de datos.")
	}
	return nil
}
