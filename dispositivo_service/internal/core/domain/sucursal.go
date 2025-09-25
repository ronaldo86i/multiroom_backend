package domain

import "time"

type Sucursal struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	Estado   string    `json:"estado,omitempty"`
	CreadoEn time.Time `json:"creadoEn"`
}

type SucursalRequest struct {
	Nombre string `json:"nombre"`
	PaisId int    `json:"paisId"`
}

type SucursalInfo struct {
	Sucursal
}

type SucursalDetail struct {
	Sucursal
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
	Pais          PaisInfo   `json:"pais"`
}

type SucursalId struct {
	Id int `json:"id"`
}
