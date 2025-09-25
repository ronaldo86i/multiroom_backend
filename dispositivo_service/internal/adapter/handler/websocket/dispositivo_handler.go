package websocket

import (
	"context"
	"fmt"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/port"
	"multiroom/dispositivo-service/internal/core/util"
	"sync"

	"github.com/gofiber/contrib/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DispositivoHandlerWS struct {
	dispositivoService port.DispositivoService
	salaService        port.SalaService
	rabbitService      port.RabbitMQService
}

// ---------- CONTROL DE CONSUMIDORES ÚNICOS ----------
var oncePerQueue sync.Map // map[string]*sync.Once

func getOnceForQueue(queue string) *sync.Once {
	onceIface, _ := oncePerQueue.LoadOrStore(queue, &sync.Once{})
	return onceIface.(*sync.Once)
}

func publishSalaAsync(rabbitMQ port.RabbitMQService, sala domain.SalaDetail) {
	err := rabbitMQ.Publish("salas", sala, amqp.Table{
		// Máximo de mensajes
		amqp.QueueMaxLenArg: int32(1),

		// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
		amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
	})
	if err != nil {
		log.Print("Error al publicar salas:", err)
	}
	err = rabbitMQ.Publish(fmt.Sprintf("sala_%d", sala.Sala.Id), sala, amqp.Table{
		// Máximo de mensajes
		amqp.QueueMaxLenArg: int32(1),

		// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
		amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
	})
	if err != nil {
		log.Print("Error al publicar "+fmt.Sprintf("sala_%d", sala.Sala.Id)+":", err)
	}
}

func (d DispositivoHandlerWS) NotificarDispositivoHabilitar(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	dispositivoId := fmt.Sprintf("%s", c.Locals(util.ContextDispositivoIdKey).(string))
	enLinea := true
	// Obtener dispositivo
	dispositivo, err := d.dispositivoService.ObtenerDispositivoByDispositivoId(context.Background(), &dispositivoId)
	if err != nil {
		log.Println("❌ Error al buscar dispositivo:", err)
		return
	}

	queueName := fmt.Sprintf("dispositivo_%d_usuario_%s", dispositivo.Id, userId)
	log.Println("🛰️ Cliente conectado:", userId, "Queue:", queueName)
	// Guardar conexión en manager
	wsUsuariosManagers.addConnection(queueName, c)
	err = d.dispositivoService.ActualizarDispositivoEnLinea(context.Background(), &dispositivo.Id, &enLinea)
	if err != nil {
		log.Printf("Error al actualizar dispositivo: %v", err)
		return
	}
	sala, err := d.salaService.ObtenerSalaByDispositivoId(context.Background(), &dispositivoId)
	if err != nil {
		return
	}
	if sala != nil {
		// Publicar estado actualizado de la sala
		publishSalaAsync(d.rabbitService, *sala)
	}
	// Consumidor único por cola
	getOnceForQueue(queueName).Do(func() {
		go func() {
			log.Printf("📡 Iniciando consumidor único para cola '%s'", queueName)
			err := d.rabbitService.StartConsumer(queueName, func(msg amqp.Delivery) {
				conns := wsUsuariosManagers.loadConnections(queueName)
				if conns == nil {
					_ = msg.Ack(false)
					return
				}

				conns.Range(func(key, _ any) bool {
					conn, ok := key.(*websocket.Conn)
					if !ok {
						log.Printf("❌ Tipo inesperado en conexiones para queue %s", queueName)
						return true
					}

					// Enviar mensaje de manera concurrente para no bloquear otras conexiones
					go func(c *websocket.Conn, data []byte) {
						if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
							log.Printf("❌ Error enviando WebSocket a %s: %v", queueName, err)
							wsUsuariosManagers.removeConnection(queueName, c)
							_ = c.Close()
						}
					}(conn, msg.Body)

					return true
				})

				_ = msg.Ack(false)
			}, amqp.Table{
				// Máximo de mensajes
				amqp.QueueMaxLenArg: int32(1),
				// Política de descarte
				amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
			})
			if err != nil {
				log.Printf("❌ Error iniciando consumidor para %s: %v", queueName, err)
			}
		}()
	})

	// Limpiar al desconectar
	defer func() {
		enLinea := false
		err := d.dispositivoService.ActualizarDispositivoEnLinea(context.Background(), &dispositivo.Id, &enLinea)
		if err != nil {
			log.Printf("Error al actualizar dispositivo: %v", err)
		}
		sala, err := d.salaService.ObtenerSalaByDispositivoId(context.Background(), &dispositivoId)
		if err != nil {
			log.Printf(err.Error())
		} else if sala != nil {
			// Publicar estado actualizado de la sala
			publishSalaAsync(d.rabbitService, *sala)
		}

		// Remover conexión del manager y cerrar WS
		wsUsuariosManagers.removeConnection(queueName, c)
		_ = c.Close()
		log.Println("❌ Cliente desconectado:", queueName)

	}()

	// Leer mensajes desde el cliente (si aplica)
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("🔌 Cliente %v desconectado: %v", queueName, err)
			break
		}
		log.Printf("📥 Mensaje desde cliente %v: %s", queueName, msg)
	}
}

func NewDispositivoHandlerWS(dispositivoService port.DispositivoService, salaService port.SalaService, rabbitService port.RabbitMQService) *DispositivoHandlerWS {
	return &DispositivoHandlerWS{dispositivoService: dispositivoService, salaService: salaService, rabbitService: rabbitService}
}

var _ port.DispositivoHandlerWS = (*DispositivoHandlerWS)(nil)
