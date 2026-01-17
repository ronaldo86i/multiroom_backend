package domain

import "time"

type SucursalId struct {
	Id int `json:"id"`
}

type SucursalInfo struct {
	SucursalId
	Nombre   string    `json:"nombre"`
	Estado   string    `json:"estado,omitempty"`
	CreadoEn time.Time `json:"creadoEn"`
}

type SucursalRequest struct {
	Nombre string `json:"nombre"`
	PaisId int    `json:"paisId"`
}

type Sucursal struct {
	SucursalInfo
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
	Pais          PaisInfo   `json:"pais"`
}
