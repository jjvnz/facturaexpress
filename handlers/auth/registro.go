package handlers

import (
	"facturaexpress/common"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register maneja el registro de usuarios.
func Register(c *gin.Context, db *data.DB) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("JSON_BINDING_FAILED", "Error al procesar los datos del usuario."))
		return
	}

	if err := CheckUsernameEmail(db, user.Username, user.Email); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("PASSWORD_HASHING_FAILED", "Error al hashear la contraseña."))
		return
	}

	userID, err := saveUser(db, user.Username, hashedPassword, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := saveUserRole(db, userID); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado con éxito."})
}

// CheckUsernameEmail verifica si el nombre de usuario o el correo electrónico ya están en uso.
func CheckUsernameEmail(db *data.DB, username string, email string) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM usuarios WHERE nombre_usuario = $1 OR correo = $2")
	if err != nil {
		return models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
	}
	defer stmt.Close()
	row := stmt.QueryRow(username, email)
	var count int
	if err = row.Scan(&count); err != nil || count > 0 {
		return models.ErrorResponseInit("USERNAME_OR_EMAIL_IN_USE", "El nombre de usuario o el correo electrónico ya están en uso.")
	}
	return nil
}

// saveUser guarda al usuario en la base de datos y devuelve su ID.
func saveUser(db *data.DB, username string, hashedPassword []byte, email string) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO usuarios (nombre_usuario, password, correo) VALUES ($1, $2, $3) RETURNING id`)
	if err != nil {
		return 0, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
	}
	defer stmt.Close()
	var userID int64
	if err = stmt.QueryRow(username, hashedPassword, email).Scan(&userID); err != nil {
		return 0, models.ErrorResponseInit("DATABASE_SAVE_FAILED", "Error al guardar en la base de datos.")
	}
	return userID, nil
}

// saveUserRole guarda el rol del usuario en la base de datos.
func saveUserRole(db *data.DB, userID int64) error {
	stmt, err := db.Prepare(`SELECT id FROM roles WHERE name=$1`)
	if err != nil {
		return models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
	}
	defer stmt.Close()
	row := stmt.QueryRow(common.USER)
	var roleID int64
	if err = row.Scan(&roleID); err != nil {
		return models.ErrorResponseInit("ROLE_ID_RETRIEVAL_FAILED", "Error al obtener el ID del rol.")
	}

	stmt, err = db.Prepare(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`)
	if err != nil {
		return models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta.")
	}
	defer stmt.Close()
	if _, err = stmt.Exec(userID, roleID); err != nil {
		return models.ErrorResponseInit("DATABASE_SAVE_FAILED", "Error al guardar en la base de datos.")
	}
	return nil
}
