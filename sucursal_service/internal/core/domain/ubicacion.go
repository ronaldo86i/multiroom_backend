package domain

type UbicacionRequest struct {
	Nombre         string `json:"nombre"`
	Estado         string `json:"estado"`
	EsVendible     bool   `json:"esVendible"`
	PrioridadVenta int    `json:"prioridadVenta"`
	SucursalId     int    `json:"sucursalId"`
}

type Ubicacion struct {
	Id             int           `json:"id"`
	Nombre         string        `json:"nombre"`
	Estado         string        `json:"estado"`
	EsVendible     bool          `json:"esVendible"`
	PrioridadVenta int           `json:"prioridadVenta"`
	Sucursal       *SucursalInfo `json:"sucursal,omitempty"`
}

type UbicacionId struct {
	Id int `json:"id"`
}
