package domain

import "time"

type UsuarioSucursalRequest struct {
	SucursalId int    `json:"sucursalId"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type UsuarioSucursalInfo struct {
	Id       int       `json:"id"`
	Username string    `json:"username"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
	Sucursal Sucursal  `json:"sucursal"`
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

type LoginSucursalRequest struct {
	SucursalId int    `json:"sucursalId"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}
