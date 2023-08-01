package models

import "github.com/golang-jwt/jwt"

type Claims struct {
	UsuarioID int64  `json:"usuario_id"`
	Role      string `json:"role"`
	jwt.StandardClaims
}
