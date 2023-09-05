package helpers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
)

func SaveUser(db *data.PostgresAdapter, username string, hashedPassword []byte, email string) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id`)
	if err != nil {
		return 0, models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta.")
	}
	defer stmt.Close()
	var userID int64
	if err = stmt.QueryRow(username, hashedPassword, email).Scan(&userID); err != nil {
		return 0, models.ErrorResponseInit(common.ErrDatabaseSaveFailed, "Error al guardar en la base de datos.")
	}
	return userID, nil
}
