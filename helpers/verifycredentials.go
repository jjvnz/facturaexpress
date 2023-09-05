package helpers

import (
	"database/sql"
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"

	"golang.org/x/crypto/bcrypt"
)

func VerifyCredentials(db *data.PostgresAdapter, correo string, password string) (models.User, error) {
	var user models.User
	stmt, err := db.Prepare(`SELECT usuarios.id ,usuarios.nombre_usuario ,usuarios.password ,roles.name FROM usuarios INNER JOIN user_roles ON usuarios.id = user_roles.user_id INNER JOIN roles ON user_roles.role_id = roles.id WHERE usuarios.correo=$1`)
	if err != nil {
		return user, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(correo)
	err = row.Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err == sql.ErrNoRows {
		return user, err
	} else if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return user, models.ErrorResponseInit(common.ErrIncorrectPassword, "La contraseña que ingresaste es incorrecta.")
	}
	return user, nil
}
