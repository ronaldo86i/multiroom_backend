package domain

import "time"

type PermisoId struct {
	Id int `json:"id"`
}
type Permiso struct {
	PermisoId
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Icono       *string   `json:"icono"`
	CreadoEn    time.Time `json:"creadoEn"`
}
