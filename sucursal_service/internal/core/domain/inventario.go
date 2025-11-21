package domain

import "time"

type Inventario struct {
	Id        int       `json:"id"`
	Producto  Producto  `json:"producto"`
	Ubicacion Ubicacion `json:"ubicacion"`
	Stock     int64     `json:"stock"`
}

// Definimos las "direcciones" del ajuste
type ajusteDirection int

const (
	AjusteSalida  ajusteDirection = -1 // Solo permite cantidades negativas
	AjusteEntrada ajusteDirection = 1  // Solo permite cantidades positivas
	AjusteMixto   ajusteDirection = 0  // Permite ambas (para conteo físico)
)

// ValidAjusteTypes Define la dirección permitida para cada tipo de ajuste
var ValidAjusteTypes = map[string]ajusteDirection{
	// Tipos de Salida
	"VENCIMIENTO":          AjusteSalida,
	"MERMA":                AjusteSalida,
	"CONSUMO_INTERNO":      AjusteSalida,
	"ROBO_HURTO":           AjusteSalida,
	"DEVOLUCION_PROVEEDOR": AjusteSalida,

	// Tipos de Entrada
	"CARGA_INICIAL":        AjusteEntrada,
	"REINGRESO_SIN_COMPRA": AjusteEntrada,

	// Tipos Mixtos ( +/- )
	"ERROR_CONTEO": AjusteMixto,
}

type AjusteId struct {
	Id int `json:"id"`
}

type AjusteInventarioRequest struct {
	TipoAjuste string                           `json:"tipoAjuste"`
	Motivo     string                           `json:"motivo"`
	SucursalId int                              `json:"sucursalId"`
	UsuarioId  int                              `json:"usuarioId"`
	Detalles   []DetalleAjusteInventarioRequest `json:"detalles"`
}

type DetalleAjusteInventarioRequest struct {
	ProductoId  int   `json:"productoId"`
	Cantidad    int64 `json:"cantidad"`
	UbicacionId int   `json:"ubicacionId"`
}

type AjusteInventarioInfo struct {
	AjusteId
	TipoAjuste string        `json:"tipoAjuste"`
	Motivo     string        `json:"motivo"`
	Usuario    UsuarioSimple `json:"usuario"`
	Sucursal   SucursalInfo  `json:"sucursal"`
	//Detalles   []DetalleAjusteInventario `json:"detalles"`
}

type AjusteInventario struct {
	AjusteInventarioInfo
	Detalles []DetalleAjusteInventario `json:"detalles"`
}
type DetalleAjusteInventario struct {
	Id        int       `json:"id"`
	Producto  Producto  `json:"producto"`
	Ubicacion Ubicacion `json:"ubicacion"`
	Cantidad  int64     `json:"cantidad"`
}

type TransferenciaId struct {
	Id int `json:"id"`
}
type TransferenciaRequest struct {
	UbicacionOrigenId  int                           `json:"ubicacionOrigenId"`
	UbicacionDestinoId int                           `json:"ubicacionDestinoId"`
	UsuarioId          int                           `json:"usuarioId"`
	Motivo             string                        `json:"motivo"`
	Detalles           []DetalleTransferenciaRequest `json:"detalles"`
}

type DetalleTransferenciaRequest struct {
	ProductoId int   `json:"productoId"`
	Cantidad   int64 `json:"cantidad"`
}

type TransferenciaInventarioInfo struct {
	TransferenciaId
	UbicacionOrigen  Ubicacion     `json:"ubicacionOrigen"`
	UbicacionDestino Ubicacion     `json:"ubicacionDestino"`
	Usuario          UsuarioSimple `json:"usuario"`
	Motivo           string        `json:"motivo"`
	Fecha            time.Time     `json:"fecha"`
}
type TransferenciaInventario struct {
	TransferenciaInventarioInfo
	Detalles []DetalleTransferenciaInventario `json:"detalles"`
}
type DetalleTransferenciaInventario struct {
	Id       int      `json:"id"`
	Producto Producto `json:"producto"`
	Cantidad int64    `json:"cantidad"`
}
