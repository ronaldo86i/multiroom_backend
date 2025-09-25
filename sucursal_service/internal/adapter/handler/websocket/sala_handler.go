package websocket

import (
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/port"
	"sync"

	"github.com/gofiber/contrib/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SalaHandlerWS struct {
	salaService     port.SalaService
	rabbitMQService port.RabbitMQService
}

// ---------- CONTROL DE CONSUMIDORES ÚNICOS ----------
var oncePerQueue sync.Map // map[string]*sync.Once

func getOnceForQueue(queue string) *sync.Once {
	onceIface, _ := oncePerQueue.LoadOrStore(queue, &sync.Once{})
	return onceIface.(*sync.Once)
}

// ---------- USO POR SUCURSAL ----------
func (s SalaHandlerWS) UsoSalasBySucursalId(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	sucursalIdLocal := c.Locals("sucursalId").(int)
	sucursalIdStr := fmt.Sprintf("%d", sucursalIdLocal)

	log.Println("🛰️ Usuario conectado:", userId, "a sucursal:", sucursalIdStr)

	wsUsuariosSucursalManagers.addConnection(sucursalIdStr, c)
	queueName := fmt.Sprintf("sucursal_%s_salas", sucursalIdStr)

	// Consumidor único por sucursal
	getOnceForQueue(queueName).Do(func() {
		go func() {
			log.Printf("📡 Iniciando consumidor único para cola '%s'", queueName)
			err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
				conns, ok := wsUsuariosSucursalManagers.getConnections(sucursalIdStr)
				if !ok {
					_ = msg.Ack(false)
					return
				}

				conns.Range(func(key, _ any) bool {
					conn := key.(*websocket.Conn)
					if err := conn.WriteMessage(websocket.TextMessage, msg.Body); err != nil {
						log.Printf("❌ Error WS en sucursal %s: %v", sucursalIdStr, err)
						wsUsuariosSucursalManagers.removeConnection(sucursalIdStr, conn)
						_ = conn.Close()
					}
					return true
				})
				_ = msg.Ack(false)
			}, amqp.Table{
				amqp.QueueMaxLenArg:   int32(1),
				amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
			})
			if err != nil {
				log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
			}
		}()
	})

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

// ---------- USO GENERAL ----------
func (s SalaHandlerWS) UsoSalas(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	log.Println("🛰️ Usuario conectado:", userId)

	wsUsuariosManagers.addConnection(userId, c)
	queueName := "salas"

	// Consumidor único para "salas"
	getOnceForQueue(queueName).Do(func() {
		go func() {
			log.Printf("📡 Iniciando consumidor único para cola '%s'", queueName)
			err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
				val, ok := wsUsuariosManagers.Load(userId)
				if !ok {
					_ = msg.Nack(false, true)
					return
				}
				conns := val.(*sync.Map)

				conns.Range(func(key, _ any) bool {
					conn := key.(*websocket.Conn)
					if err := conn.WriteMessage(websocket.TextMessage, msg.Body); err != nil {
						log.Printf("❌ Error enviando WebSocket al usuario %s: %v", userId, err)
						wsUsuariosManagers.removeConnection(userId, conn)
						_ = conn.Close()
					}
					return true
				})
				_ = msg.Ack(false)
			}, amqp.Table{
				amqp.QueueMaxLenArg:   int32(1),
				amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
			})
			if err != nil {
				log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
			}
		}()
	})

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

// ---------- USO POR SALA ----------
func (s SalaHandlerWS) UsoSala(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	salaId := c.Params("salaId")
	log.Println("🛰️ Usuario conectado:", userId, "a sala:", salaId)

	wsUsuariosBySalaManagers.addConnection(salaId, c)
	queueName := "salas_" + salaId

	// Consumidor único por sala
	getOnceForQueue(queueName).Do(func() {
		go func() {
			log.Printf("📡 Iniciando consumidor único para cola '%s'", queueName)
			err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
				val, ok := wsUsuariosBySalaManagers.Load(salaId)
				if !ok {
					_ = msg.Ack(false)
					return
				}

				conns := val.(*sync.Map)
				conns.Range(func(key, _ any) bool {
					conn := key.(*websocket.Conn)
					if err := conn.WriteMessage(websocket.TextMessage, msg.Body); err != nil {
						log.Printf("❌ Error WS en sala %s: %v", salaId, err)
						wsUsuariosBySalaManagers.removeConnection(salaId, conn)
						_ = conn.Close()
					}
					return true
				})
				_ = msg.Ack(false)
			}, amqp.Table{
				amqp.QueueMaxLenArg:   int32(1),
				amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
			})
			if err != nil {
				log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
			}
		}()
	})

	defer func() {
		wsUsuariosBySalaManagers.removeConnection(salaId, c)
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

// ---------- CONSTRUCTOR ----------
func NewSalaHandlerWS(salaService port.SalaService, rabbitMQService port.RabbitMQService) *SalaHandlerWS {
	return &SalaHandlerWS{salaService: salaService, rabbitMQService: rabbitMQService}
}

var _ port.SalaHandlerWS = (*SalaHandlerWS)(nil)
