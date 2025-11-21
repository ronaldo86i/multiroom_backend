package domain

import "time"

type Producto struct {
	Id            int        `json:"id"`
	Nombre        string     `json:"nombre"`
	Estado        string     `json:"estado"`
	UrlFoto       string     `json:"urlFoto,omitempty"`
	Precio        float64    `json:"precio"`
	CreadoEn      time.Time  `json:"creadoEn"`
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
}

type ProductoRequest struct {
	Nombre string  `json:"nombre"`
	Estado string  `json:"estado"`
	Precio float64 `json:"precio"`
}

type ProductoId struct {
	Id int `json:"id"`
}
type ProductosUsoSalaRequest struct {
	Efectivo  float64           `json:"efectivo"`
	QR        float64           `json:"qr"`
	Tarjeta   float64           `json:"tarjeta"`
	Productos []ProductoUsoSala `json:"productos"`
}
type ProductoUsoSala struct {
	ProductoId string `json:"productoId"`
	Cantidad   uint64 `json:"cantidad"`
}

type Movimiento struct {
	ProductoId  string    `json:"productoId"`
	Cantidad    uint64    `json:"cantidad"`
	Tipo        string    `json:"tipo"`
	Fecha       time.Time `json:"fecha"`
	Descripcion string    `json:"descripcion,omitempty"`
	Username    string    `json:"username"`
	UsuarioId   int       `json:"-"`
}
