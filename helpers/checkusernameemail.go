package helpers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
)

func CheckUsernameEmail(db *data.PostgresAdapter, username string, email string) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM usuarios WHERE nombre_usuario = $1 OR correo = $2")
	if err != nil {
		return models.ErrorResponseInit(common.ErrQueryPreparationFailed, "Error al preparar la consulta.")
	}
	defer stmt.Close()
	row := stmt.QueryRow(username, email)
	var count int
	if err = row.Scan(&count); err != nil || count > 0 {
		return models.ErrorResponseInit("USERNAME_OR_EMAIL_IN_USE", "El nombre de usuario o el correo electrónico ya están en uso.")
	}
	return nil
}
