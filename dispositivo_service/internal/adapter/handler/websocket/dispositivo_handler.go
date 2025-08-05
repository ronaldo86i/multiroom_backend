package websocket

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"multiroom/dispositivo-service/internal/core/port"
	"sync"
)

type DispositivoHandlerWS struct {
	dispositivoService port.DispositivoService
	rabbitService      port.RabbitMQService
}

func (d DispositivoHandlerWS) NotificarDispositivoHabilitar(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	log.Println("🛰️ Cliente conectado:", userId)

	addConnection(userId, c)
	queueName := "dispositivo_usuario_" + userId

	go func() {
		log.Printf("📡 Iniciando consumidor para cola '%s'", queueName)
		err := d.rabbitService.StartConsumer(queueName, func(msg amqp.Delivery) {
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
					removeConnection(userId, conn)
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
		removeConnection(userId, c)
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

func NewDispositivoHandlerWS(dispositivoService port.DispositivoService, rabbitService port.RabbitMQService) *DispositivoHandlerWS {
	return &DispositivoHandlerWS{dispositivoService: dispositivoService, rabbitService: rabbitService}
}

var _ port.DispositivoHandlerWS = (*DispositivoHandlerWS)(nil)
