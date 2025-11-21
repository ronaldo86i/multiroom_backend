package domain

type MetodoPago struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
	Estado string `json:"estado"`
}
