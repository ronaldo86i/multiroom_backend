package domain

import "time"

type Rol struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
}

type RolInfo struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
}
