package models

type Usuario struct {
	ID            int64  `json:"id"`
	NombreUsuario string `json:"nombre_usuario"`
	Password      string `json:"password"`
	Correo        string `json:"correo"`
	Role          string `json:"role"`
}

type LoginData struct {
	Correo   string `json:"correo" binding:"required"`
	Password string `json:"password" binding:"required"`
}
