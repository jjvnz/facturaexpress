package handlers

import (
	"database/sql"
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AssignRole(c *gin.Context) {
	userID := c.Param("id")
	newRoleID := c.Param("newRoleID")

	if userID == "" || newRoleID == " " {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error en los paramentros de consulta"})
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del usuario debe ser un número entero válido"})
		return
	}
	roleIDInt, err := strconv.Atoi(newRoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del rol debe ser un número entero válido"})
		return
	}

	db := data.GetInstance()

	var count int
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el usuario existe"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El usuario especificado no existe"})
		return
	}
	stmt, err = db.Prepare(`SELECT COUNT(*) FROM roles WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el rol existe"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El rol especificado no existe"})
		return
	}

	stmt, err = db.Prepare(`SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt, roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el usuario ya tiene asignado el rol especificado"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El usuario ya tiene asignado el rol especificado"})
		return
	}

	stmt, err = db.Prepare(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(userIDInt, roleIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al asignar el rol al usuario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rol asignado con éxito",
	})
}

func ListRoles(c *gin.Context) {
	db := data.GetInstance()

	rows, err := db.Query(`SELECT id, name FROM roles`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la lista de roles"})
		return
	}
	defer rows.Close()
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al escanear la fila de la base de datos"})
			return
		}
		roles = append(roles, role)
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

func UpdateRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleID")

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_USER_ID", "El ID del usuario debe ser un número entero válido"))
		return
	}
	roleIDInt, err := strconv.Atoi(roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("INVALID_ROLE_ID", "El ID del nuevo rol debe ser un número entero válido"))
		return
	}

	db := data.GetInstance()

	var count int
	stmt, err := db.Prepare(`SELECT COUNT(*) FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("USER_VERIFICATION_FAILED", "Error al verificar si el usuario existe"))
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit("USER_NOT_FOUND", "El usuario especificado no existe"))
		return
	}
	stmt, err = db.Prepare(`SELECT COUNT(*) FROM roles WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("ROLE_VERIFICATION_FAILED", "Error al verificar si el rol existe"))
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseInit("ROLE_NOT_FOUND", "El rol especificado no existe"))
		return
	}

	stmt, err = db.Prepare(`SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userIDInt, roleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("USER_ROLE_VERIFICATION_FAILED", "Error al verificar si el usuario ya tiene el rol especificado"))
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseInit("USER_ALREADY_HAS_ROLE", "El usuario ya tiene el rol especificado. Por favor, actualice a otro rol."))
		return
	}

	stmt, err = db.Prepare(`UPDATE user_roles SET role_id = $1 WHERE user_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(roleIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("USER_ROLE_UPDATE_FAILED", "Error al actualizar el rol del usuario"))
		return
	}

	stmt, err = db.Prepare(`SELECT jwt_token FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta"))
		return
	}
	defer stmt.Close()
	var tokenString sql.NullString
	err = stmt.QueryRow(userIDInt).Scan(&tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("JWT_TOKEN_RETRIEVAL_FAILED", "Error al obtener el token JWT del usuario"))
		return
	}
	if tokenString.Valid && tokenString.String != "" {
		stmt, err := db.Prepare(`INSERT INTO jwt_blacklist (token) VALUES ($1)`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("QUERY_PREPARATION_FAILED", "Error al preparar la consulta"))
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(tokenString.String)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseInit("JWT_TOKEN_BLACKLISTING_FAILED", "Error al agregar el token a la lista negra"))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rol actualizado con éxito",
	})
}
