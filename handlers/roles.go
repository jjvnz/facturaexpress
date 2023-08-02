package handlers

import (
	"facturaexpress/data"
	"facturaexpress/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AssignRole(c *gin.Context, db *data.DB) {
	// Obtener el ID del usuario y el ID del rol de los parámetros de la solicitud
	userID := c.Param("userID")
	roleID := c.Param("roleID")

	if userID == "" || roleID == " " {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error en los paramentros de consulta"})
	}

	// Convertir los valores de userID y roleID a enteros
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del usuario debe ser un número entero válido"})
		return
	}
	roleIDInt, err := strconv.Atoi(roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del rol debe ser un número entero válido"})
		return
	}

	// Verificar si el usuario y el rol existen en la base de datos
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

	// Verificar si el usuario ya tiene asignado el rol especificado
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

	// Asignar el rol al usuario en la base de datos
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

	// Devolver una respuesta al cliente
	c.JSON(http.StatusOK, gin.H{
		"message": "Rol asignado con éxito",
	})
}

func ListRoles(c *gin.Context, db *data.DB) {
	// Obtener la lista de roles de la base de datos
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

	// Devolver una respuesta al cliente con la lista de roles
	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

func ActualizarRol(c *gin.Context, db *data.DB) {
	// Obtener el ID del usuario y el ID del nuevo rol de los parámetros de la solicitud
	userID := c.Param("userID")
	newRoleID := c.Param("newRoleID")

	// Convertir los valores de userID y newRoleID a enteros
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del usuario debe ser un número entero válido"})
		return
	}
	newRoleIDInt, err := strconv.Atoi(newRoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El ID del nuevo rol debe ser un número entero válido"})
		return
	}

	// Verificar si el usuario y el nuevo rol existen en la base de datos
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
	err = stmt.QueryRow(newRoleIDInt).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar si el rol existe"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El rol especificado no existe"})
		return
	}

	// Actualizar el rol del usuario en la base de datos
	stmt, err = db.Prepare(`UPDATE user_roles SET role_id = $1 WHERE user_id = $2`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(newRoleIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar el rol del usuario"})
		return
	}

	// Obtener el token JWT del usuario cuyo rol ha cambiado
	stmt, err = db.Prepare(`SELECT jwt_token FROM usuarios WHERE id = $1`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al Obtener el token JWT del usuario"})
		return
	}
	defer stmt.Close()
	var tokenString string
	err = stmt.QueryRow(userIDInt).Scan(&tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al Obtener el token JWT del usuario"})
		return
	}
	if tokenString != "" {
		stmt, err := db.Prepare(`INSERT INTO jwt_blacklist (token) VALUES ($1)`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(tokenString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al agregar el token a la lista negra"})
			return
		}
	}

	// Devolver una respuesta al cliente
	c.JSON(http.StatusOK, gin.H{
		"message": "Rol actualizado con éxito",
	})
}
