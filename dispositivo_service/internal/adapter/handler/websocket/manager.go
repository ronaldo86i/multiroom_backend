package websocket

import (
	"github.com/gofiber/contrib/websocket"
	"sync"
)

var wsUsuariosManagers = &sync.Map{} // userId → []*websocket.Conn

// Añadir conexión
func addConnection(userId string, conn *websocket.Conn) {
	val, _ := wsUsuariosManagers.LoadOrStore(userId, &sync.Map{})
	conns := val.(*sync.Map)
	conns.Store(conn, struct{}{})
}

// Eliminar conexión
func removeConnection(userId string, conn *websocket.Conn) {
	val, ok := wsUsuariosManagers.Load(userId)
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
		wsUsuariosManagers.Delete(userId)
	}
}
