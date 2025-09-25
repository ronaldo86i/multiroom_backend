package domain

import "time"

type DispositivoInfo struct {
	Id            int            `json:"id"`
	DispositivoId string         `json:"dispositivoId"`
	Nombre        string         `json:"nombre"`
	Estado        string         `json:"estado"`
	CreadoEn      time.Time      `json:"creadoEn"`
	Usuario       *UsuarioSimple `json:"usuario,omitempty"`
	EnLinea       bool           `json:"enLinea"`
}
