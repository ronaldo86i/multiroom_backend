package domain

import "time"

type ProductoInfo struct {
	Id              int        `json:"id"`
	Nombre          string     `json:"nombre"`
	Estado          string     `json:"estado"`
	UrlFoto         string     `json:"urlFoto,omitempty"`
	EsInventariable bool       `json:"esInventariable"`
	CreadoEn        time.Time  `json:"creadoEn"`
	ActualizadoEn   time.Time  `json:"actualizadoEn"`
	EliminadoEn     *time.Time `json:"eliminadoEn"`
}
type Producto struct {
	ProductoInfo
	Categoria *ProductoCategoriaInfo `json:"categoria"`
}

type ProductoRequest struct {
	Nombre          string  `json:"nombre"`
	Estado          string  `json:"estado"`
	Precio          float64 `json:"precio"`
	EsInventariable bool    `json:"esInventariable"`
	CategoriaId     *int    `json:"categoriaId"`
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

type VentaDiaria struct {
	Fecha string  `json:"fecha"`
	Total float64 `json:"total"`
}

type CompraDiaria struct {
	Fecha string  `json:"fecha"`
	Total float64 `json:"total"`
}
type ProductoStat struct {
	TotalVentas     float64        `json:"totalVentas"`
	CantidadVentas  int            `json:"cantidadVentas"`
	TotalCompras    float64        `json:"totalCompras"`
	CantidadCompras int            `json:"cantidadCompras"`
	Producto        ProductoInfo   `json:"producto"`
	VentasDiarias   []VentaDiaria  `json:"ventasDiarias"`
	ComprasDiarias  []CompraDiaria `json:"comprasDiarias"`
}

type ProductoVentaStat struct {
	Producto       ProductoInfo `json:"producto"`
	TotalVentas    float64      `json:"totalVentas"`
	CantidadVentas int          `json:"cantidadVentas"`
}

type ProductoSucursalInfo struct {
	Id       int          `json:"id"`
	Precio   float64      `json:"precio"`
	Estado   string       `json:"estado"`
	Producto ProductoInfo `json:"producto"`
}

type ProductoSucursalUpdateRequest struct {
	Precio float64 `json:"precio"`
	Estado string  `json:"estado"`
}
