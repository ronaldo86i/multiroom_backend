package domain

import (
	"sync"
	"time"
)

type DispositivoRequest struct {
	Nombre        string `json:"nombre"`
	DispositivoId string `json:"dispositivoId"`
	UsuarioId     int    `json:"usuarioId"`
}

type DispositivoInfo struct {
	Id            int            `json:"id"`
	DispositivoId string         `json:"dispositivoId"`
	Nombre        string         `json:"nombre"`
	Estado        string         `json:"estado"`
	CreadoEn      time.Time      `json:"creadoEn"`
	Usuario       *UsuarioSimple `json:"usuario,omitempty"`
	EnLinea       bool           `json:"enLinea"`
}

type DispositivoState struct {
	mu       sync.Mutex
	EnLinea  bool
	NotifyCh chan bool
}

func (d *DispositivoState) SetEnLinea(val bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Evita notificaciones redundantes
	if d.EnLinea == val {
		return
	}

	d.EnLinea = val

	// Notifica cambio de estado sin bloquear
	select {
	case d.NotifyCh <- val:
	default:
		// Si el canal está lleno, no bloquea
	}
}

func (d *DispositivoState) GetEnLinea() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.EnLinea
}

type DispositivoMensaje struct {
	Type string `json:"type"`
}
