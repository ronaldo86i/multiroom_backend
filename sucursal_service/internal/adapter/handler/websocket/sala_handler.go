package websocket

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"multiroom/sucursal-service/internal/core/port"
	"sync"
)

type SalaHandlerWS struct {
	salaService     port.SalaService
	rabbitMQService port.RabbitMQService
}

func (s SalaHandlerWS) UsoSalasBySucursalId(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	sucursalIdLocal := c.Locals("sucursalId").(int)
	sucursalIdStr := fmt.Sprintf("%d", sucursalIdLocal)

	log.Println("🛰️ Usuario conectado:", userId, "a sucursal:", sucursalIdStr)

	wsUsuariosSucursalManagers.addConnection(sucursalIdStr, c)
	queueName := fmt.Sprintf("sucursal_%s_salas", sucursalIdStr)
	log.Println(queueName)

	go func() {
		log.Printf("📡 Iniciando consumidor para cola '%s'", queueName)
		err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
			log.Printf("📨 RabbitMQ [%s]: %s", queueName, msg.Body)

			conns, ok := wsUsuariosSucursalManagers.getConnections(sucursalIdStr)
			if !ok {
				log.Printf("⚠️ No hay conexiones en sucursal %s. Descartando mensaje.", sucursalIdStr)
				_ = msg.Nack(false, false)
				return
			}

			delivered := false
			conns.Range(func(key, _ any) bool {
				conn := key.(*websocket.Conn)
				err := conn.WriteMessage(websocket.TextMessage, msg.Body)
				if err != nil {
					log.Printf("❌ Error WS en sucursal %s: %v", sucursalIdStr, err)
					wsUsuariosSucursalManagers.removeConnection(sucursalIdStr, conn)
					_ = conn.Close()
					return true
				}
				delivered = true
				return true
			})

			if !delivered {
				log.Printf("⚠️ Nadie recibió mensaje en sucursal %s", sucursalIdStr)
				_ = msg.Nack(false, false)
				return
			}
			_ = msg.Ack(false)
		})
		if err != nil {
			log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
		}
	}()

	defer func() {
		wsUsuariosSucursalManagers.removeConnection(sucursalIdStr, c)
		_ = c.Close()
		log.Println("❌ Cliente desconectado:", userId, "de sucursal:", sucursalIdStr)
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("🔌 Cliente %v desconectado: %v", userId, err)
			return
		}
		log.Printf("📥 Mensaje desde cliente %v (sucursal %s): %s", userId, sucursalIdStr, msg)
	}
}

func (s SalaHandlerWS) UsoSalas(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	log.Println("🛰️ Usuario conectado:", userId)

	wsUsuariosManagers.addConnection(userId, c)
	queueName := "salas"
	log.Println(queueName)
	go func() {
		log.Printf("📡 Iniciando consumidor para cola '%s'", queueName)
		err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
			log.Printf("📨 RabbitMQ [%s]: %s", queueName, msg.Body)

			val, ok := wsUsuariosManagers.Load(userId)
			if !ok {
				log.Printf("⚠️ Conexión no encontrada para usuario %s. Reenviando mensaje.", userId)
				_ = msg.Nack(false, true)
				return
			}
			conns := val.(*sync.Map)
			delivered := false

			conns.Range(func(key, _ any) bool {
				conn := key.(*websocket.Conn)
				err := conn.WriteMessage(websocket.TextMessage, msg.Body)
				if err != nil {
					log.Printf("❌ Error enviando WebSocket al usuario %s: %v", userId, err)
					wsUsuariosManagers.removeConnection(userId, conn)
					_ = conn.Close()
					return true
				}
				delivered = true
				return true
			})

			if !delivered {
				log.Printf("⚠️ No se entregó mensaje a ninguna conexión para %s", userId)
				_ = msg.Nack(false, true)
				return
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("⚠️ No se pudo ACK el mensaje: %v", err)
			}
		})

		if err != nil {
			log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
		}
	}()

	defer func() {
		wsUsuariosManagers.removeConnection(userId, c)
		_ = c.Close()
		log.Println("❌ Cliente desconectado:", userId)
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("🔌 Cliente %v desconectado: %v", userId, err)
			return
		}
		log.Printf("📥 Mensaje desde cliente %v: %s", userId, msg)
	}
}

func (s SalaHandlerWS) UsoSala(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	salaId := c.Params("salaId")
	log.Println("🛰️ Usuario conectado:", userId, "a sala:", salaId)

	wsUsuariosManagers.addConnection(salaId, c) // clave = salaId
	queueName := "salas_" + salaId
	log.Println(queueName)

	go func() {
		log.Printf("📡 Iniciando consumidor para cola '%s'", queueName)
		err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
			log.Printf("📨 RabbitMQ [%s]: %s", queueName, msg.Body)

			val, ok := wsUsuariosManagers.Load(salaId) // buscar por salaId
			if !ok {
				log.Printf("⚠️ No hay conexiones en sala %s", salaId)
				_ = msg.Nack(false, false)
				return
			}

			conns := val.(*sync.Map)
			delivered := false
			conns.Range(func(key, _ any) bool {
				conn := key.(*websocket.Conn)
				err := conn.WriteMessage(websocket.TextMessage, msg.Body)
				if err != nil {
					log.Printf("❌ Error WS en sala %s: %v", salaId, err)
					wsUsuariosManagers.removeConnection(salaId, conn)
					_ = conn.Close()
					return true
				}
				delivered = true
				return true
			})

			if !delivered {
				log.Printf("⚠️ Nadie recibió mensaje en sala %s", salaId)
				_ = msg.Nack(false, false)
				return
			}
			_ = msg.Ack(false)
		})
		if err != nil {
			log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
		}
	}()

	defer func() {
		wsUsuariosManagers.removeConnection(salaId, c) // clave = salaId
		_ = c.Close()
		log.Println("❌ Cliente desconectado:", userId, "de sala:", salaId)
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("🔌 Cliente %v desconectado: %v", userId, err)
			return
		}
		log.Printf("📥 Mensaje desde cliente %v (sala %s): %s", userId, salaId, msg)
	}
}

func NewSalaHandlerWS(salaService port.SalaService, rabbitMQService port.RabbitMQService) *SalaHandlerWS {
	return &SalaHandlerWS{salaService: salaService, rabbitMQService: rabbitMQService}
}

var _ port.SalaHandlerWS = (*SalaHandlerWS)(nil)
