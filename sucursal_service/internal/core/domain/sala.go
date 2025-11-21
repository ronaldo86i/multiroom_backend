package domain

import "time"

type Sala struct {
	Id            int        `json:"id"`
	Nombre        string     `json:"nombre"`
	Estado        string     `json:"estado"`
	CreadoEn      time.Time  `json:"creadoEn"`
	ActualizadoEn time.Time  `json:"actualizadoEn"`
	EliminadoEn   *time.Time `json:"eliminadoEn"`
}
type SalaRequest struct {
	Nombre        string `json:"nombre"`
	SucursalId    int    `json:"sucursalId"`
	DispositivoId int    `json:"dispositivoId"`
}

type UsoSalaRequest struct {
	SalaId    int   `json:"salaId"`
	ClienteId int   `json:"clienteId"`
	TiempoUso int64 `json:"tiempoUso"`
}
type UsoSalaId struct {
	Id int64 `json:"id"`
}

type UsoSala struct {
	UsoSalaId
	Cliente       *ClienteInfo `json:"cliente,omitempty"`
	Inicio        time.Time    `json:"inicio"`
	Fin           *time.Time   `json:"fin,omitempty"`
	PausadoEn     *time.Time   `json:"pausadoEn,omitempty"`
	DuracionPausa float64      `json:"duracionPausa"`
	TiempoUso     float64      `json:"tiempoUso"`
	Estado        string       `json:"estado"`
	CostoTiempo   float64      `json:"costoTiempo"`
}

type SalaInfo struct {
	Sala
	Dispositivo DispositivoInfo `json:"dispositivo,omitempty"`
	Uso         *UsoSala        `json:"uso,omitempty"`
}

type SalaDetail struct {
	Sala
	Sucursal    SucursalInfo    `json:"sucursal"`
	Pais        PaisInfo        `json:"pais"`
	Dispositivo DispositivoInfo `json:"dispositivo,omitempty"`
	Uso         *UsoSala        `json:"uso,omitempty"`
}
type SalaId struct {
	Id int `json:"id"`
}
