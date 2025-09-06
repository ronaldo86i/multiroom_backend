package domain

import "time"

type UsuarioAdmin struct {
	Id           int       `json:"id"`
	Username     string    `json:"username"`
	Estado       string    `json:"estado"`
	PasswordHash string    `json:"-"`
	CreadoEn     time.Time `json:"creadoEn"`
	Roles        []RolInfo `json:"roles"`
}

type UsuarioAdminInfo struct {
	Id       int       `json:"id"`
	Username string    `json:"username"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
}

type LoginAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
