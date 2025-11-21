package domain

import "time"

type CompraId struct {
	Id int `json:"id"`
}
type CompraRequest struct {
	ProveedorId int                    `json:"proveedorId"`
	UsuarioId   int                    `json:"usuarioId"`
	SucursalId  int                    `json:"sucursalId"`
	Detalles    []DetalleCompraRequest `json:"detalles"`
}

type DetalleCompraRequest struct {
	ProductoId   int     `json:"productoId"`
	Cantidad     int     `json:"cantidad"`
	PrecioCompra float64 `json:"precioCompra"`
	PrecioVenta  float64 `json:"precioVenta"`
	UbicacionId  int     `json:"ubicacionId"`
}

type CompraInfo struct {
	CompraId
	CodigoCompra  int64         `json:"codigoCompra"`
	Usuario       UsuarioSimple `json:"usuario"`
	Proveedor     Proveedor     `json:"proveedor"`
	Sucursal      SucursalInfo  `json:"sucursal"`
	Estado        string        `json:"estado"`
	CreadoEn      time.Time     `json:"creadoEn"`
	ActualizadoEn time.Time     `json:"actualizadoEn"`
	EliminadoEn   *time.Time    `json:"eliminadoEn"`
}

type Compra struct {
	CompraInfo
	Detalles []DetalleCompra `json:"detalles"`
}

type DetalleCompra struct {
	Producto     Producto  `json:"producto"`
	Cantidad     int       `json:"cantidad"`
	PrecioCompra float64   `json:"precioCompra"`
	PrecioVenta  float64   `json:"precioVenta"`
	Ubicacion    Ubicacion `json:"ubicacion"`
}
