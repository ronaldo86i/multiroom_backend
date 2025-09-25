package domain

import "time"

type Sala struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	Estado   string    `json:"estado"`
	CreadoEn time.Time `json:"creadoEn"`
}

type UsoSala struct {
	Cliente       ClienteInfo `json:"cliente"`
	Inicio        time.Time   `json:"inicio"`
	Fin           *time.Time  `json:"fin,omitempty"`
	PausadoEn     *time.Time  `json:"pausadoEn,omitempty"`
	DuracionPausa float64     `json:"duracionPausa"`
	TiempoUso     float64     `json:"tiempoUso"`
	Estado        string      `json:"estado"`
}

type SalaDetail struct {
	Sala
	ActualizadoEn time.Time       `json:"actualizadoEn"`
	EliminadoEn   *time.Time      `json:"eliminadoEn"`
	Sucursal      SucursalInfo    `json:"sucursal"`
	Pais          PaisInfo        `json:"pais"`
	Dispositivo   DispositivoInfo `json:"dispositivo,omitempty"`
	Uso           *UsoSala        `json:"uso,omitempty"`
}
