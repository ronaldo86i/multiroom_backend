package domain

import "time"

type PermisoId struct {
	Id int `json:"id"`
}
type Permiso struct {
	PermisoId
	Nombre      string    `json:"nombre"`      // ej: "venta:registrar"
	Descripcion string    `json:"descripcion"` // ej: "Permite crear nuevas ventas"
	Icono       *string   `json:"icono"`       // ej: "ShoppingCart" (Puntero porque puede ser null)
	CreadoEn    time.Time `json:"creadoEn"`
}

type PermisoRequest struct {
	// required: No puede estar vacío
	// min/max: Longitud mínima y máxima
	// lowercase: Debe ser minúscula (aunque nosotros lo forzaremos en código)
	// contains: Debe tener dos puntos (:) - Una validación básica del formato
	Nombre      string `json:"nombre" validate:"required,min=3,max=100,contains=:"`
	Descripcion string `json:"descripcion" validate:"required,min=5"`
	Icono       string `json:"icono" validate:"omitempty,max=50"` // omitempty = opcional
}
