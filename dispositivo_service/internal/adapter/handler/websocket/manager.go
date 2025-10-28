package websocket

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// SyncMap gestiona conexiones WebSocket por key (usuario/dispositivo)
type SyncMap struct {
	sync.Map // key = string, value = *sync.Map (conexiones)
}

var (
	wsUsuariosManagers = &SyncMap{}
)

// addConnection añade una conexión WebSocket para una key específica
func (s *SyncMap) addConnection(key string, conn *websocket.Conn) {
	val, _ := s.LoadOrStore(key, &sync.Map{})
	conns := val.(*sync.Map)
	conns.Store(conn, struct{}{}) // valor dummy
}

// removeConnection elimina una conexión WebSocket de una key específica
func (s *SyncMap) removeConnection(key string, conn *websocket.Conn) {
	val, ok := s.Load(key)
	if !ok {
		return
	}
	conns := val.(*sync.Map)
	conns.Delete(conn)

	// Si no quedan conexiones, eliminar la key del SyncMap principal
	isEmpty := true
	conns.Range(func(_, _ any) bool {
		isEmpty = false
		return false // corta la iteración
	})
	if isEmpty {
		s.Delete(key)
	}
}

// loadConnections devuelve todas las conexiones activas para una key
func (s *SyncMap) loadConnections(key string) *sync.Map {
	val, ok := s.Load(key)
	if !ok {
		return nil
	}
	return val.(*sync.Map)
}
