package domain

import "time"

type Usuario struct {
	Id           int       `json:"id"`
	Username     string    `json:"username"`
	Estado       string    `json:"estado"`
	PasswordHash string    `json:"-"`
	CreadoEn     time.Time `json:"creadoEn"`
}

type UsuarioSimple struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
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
