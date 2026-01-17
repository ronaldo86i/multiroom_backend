package domain

import "time"

type ProductoCategoriaId struct {
	Id int `json:"id"`
}

type ProductoCategoriaInfo struct {
	ProductoCategoriaId
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}

type ProductoCategoria struct {
	ProductoCategoriaInfo
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

type ProductoCategoriaRequest struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Estado      string `json:"estado"`
}
