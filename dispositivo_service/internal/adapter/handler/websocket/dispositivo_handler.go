package websocket

import (
	"context"
	"fmt"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/port"
	"multiroom/dispositivo-service/internal/core/util"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DispositivoHandlerWS struct {
	dispositivoService port.DispositivoService
	salaService        port.SalaService
	rabbitService      port.RabbitMQService
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
	err = rabbitMQ.Publish(fmt.Sprintf("sucursal_%d_salas", sala.Sucursal.Id), sala, amqp.Table{
		// Máximo de mensajes
		amqp.QueueMaxLenArg: int32(1),

		// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
		amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
	})

	if err != nil {
		log.Print("Error al publicar "+fmt.Sprintf("sucursal_%d_salas", sala.Sucursal.Id)+":", err)
	}
}

func (d DispositivoHandlerWS) NotificarDispositivoHabilitar(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	dispositivoId := fmt.Sprintf("%s", c.Locals(util.ContextDispositivoIdKey).(string))
	ctx := context.Background()

	// Obtener dispositivo
	dispositivo, err := d.dispositivoService.ObtenerDispositivoByDispositivoId(ctx, &dispositivoId)
	if err != nil {
		log.Println("❌ Error al buscar dispositivo:", err)
		return
	}

	queueName := fmt.Sprintf("dispositivo_%d_usuario_%s", dispositivo.Id, userId)
	log.Println("🛰️ Cliente conectado:", userId, "Queue:", queueName)

	// Guardar conexión usando queueName directamente (sin connectionKey)
	wsUsuariosManagers.addConnection(queueName, c)

	// Estado del dispositivo con canal para señalizar cierre
	state := &domain.DispositivoState{
		NotifyCh: make(chan bool, 1),
		EnLinea:  true,
	}

	// Canal para controlar el cierre de la función
	done := make(chan struct{})

	// Cleanup seguro
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			log.Println("🧹 Iniciando cleanup para:", queueName)

			state.SetEnLinea(false)
			enLinea := false

			if err := d.dispositivoService.ActualizarDispositivoEnLinea(ctx, &dispositivo.Id, &enLinea); err != nil {
				log.Printf("❌ Error al actualizar dispositivo: %v", err)
			}

			if sala, err := d.salaService.ObtenerSalaByDispositivoId(ctx, &dispositivoId); err == nil && sala != nil {
				publishSalaAsync(d.rabbitService, *sala)
			}

			wsUsuariosManagers.removeConnection(queueName, c)

			// Cerrar conexión WebSocket
			if err := c.Close(); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("⚠️ Error al cerrar conexión WS: %v", err)
			}

			close(state.NotifyCh)
			close(done)

			log.Println("✅ Cliente desconectado y limpieza completada:", queueName)
		})
	}
	defer cleanup()

	// Marcar como conectado inicialmente
	enLinea := true
	if err := d.dispositivoService.ActualizarDispositivoEnLinea(ctx, &dispositivo.Id, &enLinea); err != nil {
		log.Printf("❌ Error al actualizar dispositivo: %v", err)
		return
	}

	if sala, err := d.salaService.ObtenerSalaByDispositivoId(ctx, &dispositivoId); err == nil && sala != nil {
		publishSalaAsync(d.rabbitService, *sala)
	}

	// Listener de cambios de estado (enLinea=false)
	go func() {
		for {
			select {
			case online, ok := <-state.NotifyCh:
				if !ok {
					return
				}
				if !online {
					log.Printf("⚡ enLinea=false detectado para %s", queueName)
					// No llamar cleanup aquí, solo loguear
					// El cleanup ya se ejecutará en el defer o en el loop principal
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Consumidor RabbitMQ - Siempre llamar, el servicio maneja duplicados internamente
	log.Printf("📡 Registrando consumidor para '%s'", queueName)
	err = d.rabbitService.StartConsumer(queueName, func(msg amqp.Delivery) {
		log.Printf("📬 Mensaje recibido en %s (tamaño: %d bytes)", queueName, len(msg.Body))

		conns := wsUsuariosManagers.loadConnections(queueName)
		if conns == nil {
			log.Printf("⚠️ No hay conexiones para %s, mensaje descartado", queueName)
			_ = msg.Ack(false)
			return
		}

		var activeConnCount int
		conns.Range(func(key, _ any) bool {
			conn, ok := key.(*websocket.Conn)
			if !ok {
				return true
			}
			activeConnCount++

			go func(c *websocket.Conn, data []byte) {
				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("❌ Error enviando WS a %s: %v", queueName, err)
					// Remover esta conexión específica
					wsUsuariosManagers.Range(func(k, v any) bool {
						if str, ok := k.(string); ok {
							if connsMap, ok := v.(*sync.Map); ok {
								connsMap.Range(func(connKey, _ any) bool {
									if connKey == c {
										wsUsuariosManagers.removeConnection(str, c)
										return false
									}
									return true
								})
							}
						}
						return true
					})
					_ = c.Close()
				} else {
					log.Printf("✉️ Mensaje enviado exitosamente a conexión de %s", queueName)
				}
			}(conn, msg.Body)

			return true
		})

		if activeConnCount > 0 {
			log.Printf("📨 Mensaje procesado para %s (%d conexiones activas)", queueName, activeConnCount)
		}

		_ = msg.Ack(false)
	}, amqp.Table{
		amqp.QueueMaxLenArg:   int32(1),
		amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
	})

	if err != nil {
		log.Printf("❌ Error registrando consumidor para %s: %v", queueName, err)
	}

	// Configuración WebSocket con timeout y ping
	c.SetReadLimit(512)
	const (
		pongWait   = 60 * time.Second // Tiempo máximo esperando pong
		pingPeriod = 45 * time.Second // Enviar ping cada 45s (antes del timeout)
	)

	_ = c.SetReadDeadline(time.Now().Add(pongWait))

	// Handler para pong - refrescar deadline
	c.SetPongHandler(func(string) error {
		_ = c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Goroutine para enviar pings periódicos
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-pingTicker.C:
				if !state.GetEnLinea() {
					return
				}
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("⚠️ Error enviando ping a %s: %v", queueName, err)
					cleanup()
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Loop principal de lectura
	for {
		select {
		case <-done:
			log.Printf("🛑 Función terminada para %s", queueName)
			return
		default:
			var msg domain.DispositivoMensaje
			if err := c.ReadJSON(&msg); err != nil {
				// Filtrar el log según el tipo de error
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("👋 Cliente %v cerró conexión normalmente", queueName)
				} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("⚠️ Cliente %v desconectado inesperadamente: %v", queueName, err)
				} else {
					log.Printf("🔌 Cliente %v desconectado: %v", queueName, err)
				}
				cleanup()
				return
			}

			// Cada mensaje recibido reinicia el timeout
			_ = c.SetReadDeadline(time.Now().Add(pongWait))

			if msg.Type == "ping" {
				state.SetEnLinea(true)
				continue
			}

			log.Printf("📩 Mensaje no manejado (%s): %+v", msg.Type, msg)
		}
	}
}

func NewDispositivoHandlerWS(dispositivoService port.DispositivoService, salaService port.SalaService, rabbitService port.RabbitMQService) *DispositivoHandlerWS {
	return &DispositivoHandlerWS{dispositivoService: dispositivoService, salaService: salaService, rabbitService: rabbitService}
}

var _ port.DispositivoHandlerWS = (*DispositivoHandlerWS)(nil)
