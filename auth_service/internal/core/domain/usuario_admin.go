package domain

import "time"

type UsuarioAdminId struct {
	Id int `json:"id"`
}
type UsuarioAdminInfo struct {
	UsuarioAdminId
	Username     string    `json:"username"`
	Estado       string    `json:"estado"`
	PasswordHash string    `json:"-"`
	CreadoEn     time.Time `json:"creadoEn"`
}

type UsuarioAdmin struct {
	UsuarioAdminInfo
	Roles      []RolInfo  `json:"roles"`
	Sucursales []Sucursal `json:"sucursales"`
	Permisos   []Permiso  `json:"permisos"`
}

type UsuarioAdminRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Estado        string `json:"estado"`
	RolesIds      []int  `json:"rolesIds"`
	SucursalesIds []int  `json:"sucursalesIds"`
}

type LoginAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
