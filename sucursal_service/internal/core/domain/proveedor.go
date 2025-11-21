package domain

import "time"

type Proveedor struct {
	Id            int        `json:"id"`
	Nombre        string     `json:"nombre"`
	Estado        string     `json:"estado"`
	Email         *string    `json:"email"`
	Telefono      *string    `json:"telefono"`
	Celular       *string    `json:"celular"`
	CreadoEn      time.Time  `json:"creadoEn"`
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
}

type ProveedorRequest struct {
	Nombre   string  `json:"nombre"`
	Estado   string  `json:"estado"`
	Email    *string `json:"email"`
	Telefono *string `json:"telefono"`
	Celular  *string `json:"celular"`
}

type ProveedorId struct {
	Id int `json:"id"`
}
