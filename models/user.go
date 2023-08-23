package models

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"nombre_usuario"`
	Password string `json:"password"`
	Email    string `json:"correo"`
	Role     string `json:"role"`
}

type LoginData struct {
	Email    string `json:"correo" binding:"required"`
	Password string `json:"password" binding:"required"`
}
