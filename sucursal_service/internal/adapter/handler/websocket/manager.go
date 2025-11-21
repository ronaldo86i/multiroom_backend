package websocket

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type SyncMap struct {
	sync.Map
}

var (
	wsUsuariosManagers         = &SyncMap{}
	wsUsuariosSucursalManagers = &SyncMap{}
	wsUsuariosBySalaManagers   = &SyncMap{}
)

// Añadir conexión
func (s *SyncMap) addConnection(key string, conn *websocket.Conn) {
	val, _ := s.LoadOrStore(key, &sync.Map{})
	conns := val.(*sync.Map)
	conns.Store(conn, struct{}{})
}

// Eliminar conexión
func (s *SyncMap) removeConnection(key string, conn *websocket.Conn) {
	val, ok := s.Load(key)
	if !ok {
		return
	}
	conns := val.(*sync.Map)
	conns.Delete(conn)
	// Opcional: si no quedan conexiones, eliminar el userId
	empty := true
	conns.Range(func(_, _ any) bool {
		empty = false
		return false
	})
	if empty {
		s.Delete(key)
	}
}

func (s *SyncMap) getConnections(key string) (*sync.Map, bool) {
	val, ok := s.Load(key)
	if !ok {
		return nil, false
	}
	return val.(*sync.Map), true
}
