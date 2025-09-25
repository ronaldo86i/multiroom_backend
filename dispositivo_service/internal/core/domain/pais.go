package domain

import "time"

type Pais struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	UrlFoto  string    `json:"urlFoto,omitempty"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
}

type PaisRequest struct {
	Nombre string `json:"nombre"`
}

type PaisInfo struct {
	Pais
}

type PaisDetail struct {
	Pais
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
}

type PaisId struct {
	Id int `json:"id"`
}
