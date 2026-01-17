package domain

import "time"

type PaisId struct {
	Id int `json:"id"`
}

type PaisInfo struct {
	PaisId
	Nombre      string    `json:"nombre"`
	CodigoLocal string    `json:"codigoLocal"`
	UrlFoto     string    `json:"urlFoto,omitempty"`
	Estado      string    `json:"estado"`
	CreadoEn    time.Time `json:"creadoEn"`
}

type Pais struct {
	PaisInfo
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
}

type PaisRequest struct {
	Nombre      string `json:"nombre"`
	CodigoLocal string `json:"codigoLocal"`
}
