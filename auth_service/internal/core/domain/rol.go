package domain

import "time"

type RolId struct {
	Id int `json:"id"`
}
type Rol struct {
	RolInfo
	Permisos []Permiso
}
type RolRequest struct {
	Nombre      string `json:"nombre"`
	Estado      string `json:"estado"`
	PermisosIds []int  `json:"permisosIds"`
}

type RolInfo struct {
	RolId
	Nombre   string    `json:"nombre"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
}
