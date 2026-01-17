package domain

import "time"

type SucursalInfo struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	CreadoEn time.Time `json:"creadoEn"`
}

type Sucursal struct {
	SucursalInfo
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
	Pais          PaisInfo   `json:"pais"`
}

type PaisInfo struct {
	Id          int       `json:"id"`
	Nombre      string    `json:"nombre"`
	CodigoLocal string    `json:"codigoLocal"`
	UrlFoto     string    `json:"urlFoto,omitempty"`
	Estado      string    `json:"estado"`
	CreadoEn    time.Time `json:"creadoEn"`
}
