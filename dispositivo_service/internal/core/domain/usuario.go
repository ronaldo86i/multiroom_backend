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
	Password string `json:"password"` // solo si es necesario
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
	Estado   string    `json:"estado,omitempty"`
	CreadoEn time.Time `json:"creadoEn"`
}

type UsuarioSimple struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}
type UsuarioResponse struct {
	Id int `json:"id"`
}

// TokenResponse contiene el token JWT devuelto al autenticarse.
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn,omitempty"` // opcional: si querés mostrar la expiración
}

type MessageData[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message"`
}

type UsuarioAdmin struct {
	Id           int       `json:"id"`
	Username     string    `json:"username"`
	Estado       string    `json:"estado"`
	PasswordHash string    `json:"-"`
	CreadoEn     time.Time `json:"creadoEn"`
	Roles        []RolInfo `json:"roles"`
}

type RolInfo struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type UsuarioSucursal struct {
	Id           int        `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Estado       string     `json:"estado"`
	CreadoEn     time.Time  `json:"creadoEn"`
	EliminadoEn  *time.Time `json:"eliminadoEn,omitempty"`
	Sucursal     Sucursal   `json:"sucursal"`
}

type Sucursal struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	Estado   string    `json:"estado,omitempty"`
	CreadoEn time.Time `json:"creadoEn"`
}
