package domain

import "time"

// LoginRequest representa la solicitud de inicio de sesión.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UsuarioRequest representa los datos necesarios para crear o actualizar un usuario.
type UsuarioRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Estado   string `json:"estado"`
}

// Usuario representa los datos completos de la tabla 'usuario'
type Usuario struct {
	Id           int       `json:"id"`
	Username     string    `json:"username"`
	Estado       string    `json:"estado"`
	PasswordHash string    `json:"-"`
	CreadoEn     time.Time `json:"creadoEn"`
}

type UsuarioInfo struct {
	Id       int       `json:"id"`
	Username string    `json:"username"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
}

type UsuarioResponse struct {
	Id int `json:"id"`
}

// TokenResponse contiene el token JWT devuelto al autenticarse.
type TokenResponse[T any] struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn,omitempty"` // opcional: si querés mostrar la expiración
	Data      T      `json:"data"`
}
