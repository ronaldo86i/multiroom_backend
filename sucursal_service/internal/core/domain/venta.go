package domain

import "time"

type VentaId struct {
	Id int `json:"id"`
}
type VentaInfo struct {
	VentaId
	CodigoVenta   int64         `json:"codigoVenta"`
	Total         float64       `json:"total"`
	Estado        string        `json:"estado"`
	CreadoEn      time.Time     `json:"creadoEn"`
	ActualizadoEn time.Time     `json:"actualizadoEn"`
	Usuario       UsuarioSimple `json:"usuario"`
	Cliente       *ClienteInfo  `json:"cliente,omitempty"`
	Sucursal      *SucursalInfo `json:"sucursal"`
	Sala          *SalaInfo     `json:"sala,omitempty"`
}

type Venta struct {
	VentaInfo
	UsoSala  *UsoSala       `json:"usoSala,omitempty"`
	Detalles []DetalleVenta `json:"detalles"`
	Pagos    *[]VentaPago   `json:"pagos"`
}

type DetalleVenta struct {
	Id          int       `json:"id"`
	Producto    Producto  `json:"producto"`
	Ubicacion   Ubicacion `json:"ubicacion"`
	Cantidad    int64     `json:"cantidad"`
	PrecioVenta float64   `json:"precioVenta"`
}

type VentaRequest struct {
	UsuarioId   int                   `json:"usuarioId"`
	SalaId      int                   `json:"salaId"`
	UsoSalaId   *int64                `json:"usoSalaId,omitempty"`
	CostoTiempo float64               `json:"costoTiempo,omitempty"`
	ClienteId   *int64                `json:"clienteId,omitempty"`
	Detalles    []DetalleVentaRequest `json:"detalles"`
}

type DetalleVentaRequest struct {
	ProductoId int   `json:"productoId"`
	Cantidad   int64 `json:"cantidad"`
}

type VentaPago struct {
	MetodoPago MetodoPago `json:"metodoPago"`
	Monto      float64    `json:"monto"`
	Referencia *string    `json:"referencia,omitempty"`
}

type RegistrarPagosRequest struct {
	Pagos []PagoRequest `json:"pagos"`
}

type PagoRequest struct {
	MetodoPagoId int     `json:"metodoPagoId"`
	Monto        float64 `json:"monto"`
	Referencia   *string `json:"referencia,omitempty"`
}
