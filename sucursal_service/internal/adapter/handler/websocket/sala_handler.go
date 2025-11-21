package websocket

import (
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/port"
	"sync"
	"sync/atomic"

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

// ---------- HELPER: GESTIÓN DE CONEXIONES WEBSOCKET ----------
type connectionManager struct {
	active      *atomic.Bool
	done        chan struct{}
	cleanup     func()
	cleanupOnce sync.Once
}

func newConnectionManager(cleanupFn func()) *connectionManager {
	cm := &connectionManager{
		active:  &atomic.Bool{},
		done:    make(chan struct{}),
		cleanup: cleanupFn,
	}
	cm.active.Store(true)
	return cm
}

func (cm *connectionManager) isActive() bool {
	return cm.active.Load()
}

func (cm *connectionManager) close() {
	cm.cleanupOnce.Do(func() {
		cm.active.Store(false)
		cm.cleanup()
		close(cm.done)
	})
}

// ---------- TIPOS DE GESTORES DE CONEXIÓN ----------
type connectionManagerType int

const (
	managerTypeUsuarios connectionManagerType = iota
	managerTypeSucursal
	managerTypeSala
)

// ---------- CONSUMIDOR GENÉRICO ----------
func (s *SalaHandlerWS) startQueueConsumer(queueName string, managerType connectionManagerType, key string, cm *connectionManager) {
	getOnceForQueue(queueName).Do(func() {
		go func() {
			log.Printf("📡 Iniciando consumidor único para cola '%s'", queueName)

			err := s.rabbitMQService.StartConsumer(queueName, func(msg amqp.Delivery) {
				if !cm.isActive() {
					_ = msg.Ack(false)
					return
				}

				// Obtener conexiones según el tipo de manager
				var conns *sync.Map
				var ok bool

				switch managerType {
				case managerTypeUsuarios:
					val, exists := wsUsuariosManagers.Load(key)
					if !exists {
						_ = msg.Ack(false)
						return
					}
					conns, ok = val.(*sync.Map)
					if !ok {
						_ = msg.Ack(false)
						return
					}

				case managerTypeSucursal:
					conns, ok = wsUsuariosSucursalManagers.getConnections(key)
					if !ok {
						_ = msg.Ack(false)
						return
					}

				case managerTypeSala:
					val, exists := wsUsuariosBySalaManagers.Load(key)
					if !exists {
						_ = msg.Ack(false)
						return
					}
					conns, ok = val.(*sync.Map)
					if !ok {
						_ = msg.Ack(false)
						return
					}

				default:
					_ = msg.Ack(false)
					return
				}

				// Enviar mensaje a todas las conexiones
				conns.Range(func(k, _ any) bool {
					conn, ok := k.(*websocket.Conn)
					if !ok {
						return true
					}

					go func(c *websocket.Conn, data []byte) {
						if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
							log.Printf("❌ Error enviando WS a %s: %v", queueName, err)

							// Remover conexión según el tipo de manager
							switch managerType {
							case managerTypeUsuarios:
								wsUsuariosManagers.removeConnection(key, c)
							case managerTypeSucursal:
								wsUsuariosSucursalManagers.removeConnection(key, c)
							case managerTypeSala:
								wsUsuariosBySalaManagers.removeConnection(key, c)
							}

							_ = c.Close()
							cm.active.Store(false)
						}
					}(conn, msg.Body)

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
}

// ---------- LOOP DE LECTURA GENÉRICO ----------
func (s *SalaHandlerWS) readLoop(c *websocket.Conn, cm *connectionManager, contextInfo string) {
	c.SetReadLimit(512)

	for {
		select {
		case <-cm.done:
			log.Printf("🛑 Función terminada para %s", contextInfo)
			return
		default:
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Printf("🔌 Cliente desconectado: %v", err)
				cm.close()
				return
			}

			if len(msg) > 0 {
				log.Printf("📥 Mensaje desde %s: %s", contextInfo, msg)
			}
		}
	}
}

// ---------- USO POR SUCURSAL ----------
func (s SalaHandlerWS) UsoSalasBySucursalId(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	sucursalIdLocal := c.Locals("sucursalId").(int)
	sucursalIdStr := fmt.Sprintf("%d", sucursalIdLocal)
	queueName := fmt.Sprintf("sucursal_%s_salas", sucursalIdStr)
	// Key única: incluye puntero de la conexión para que cada dispositivo sea único
	connectionKey := fmt.Sprintf("sucursal_%s_%s_%p", sucursalIdStr, userId, c)

	log.Println("🛰️ Usuario conectado:", userId, "a sucursal:", sucursalIdStr)

	wsUsuariosSucursalManagers.addConnection(connectionKey, c)

	cm := newConnectionManager(func() {
		log.Println("🧹 Iniciando cleanup para sucursal:", sucursalIdStr)
		wsUsuariosSucursalManagers.removeConnection(connectionKey, c)

		if err := c.Close(); err != nil {
			log.Printf("⚠️ Error al cerrar conexión WS: %v", err)
		}

		log.Println("❌ Cliente desconectado y limpieza completada - Usuario:", userId, "Sucursal:", sucursalIdStr)
	})
	defer cm.close()

	s.startQueueConsumer(queueName, managerTypeSucursal, connectionKey, cm)
	s.readLoop(c, cm, fmt.Sprintf("usuario %s (sucursal %s)", userId, sucursalIdStr))
}

// ---------- USO GENERAL ----------
func (s SalaHandlerWS) UsoSalas(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	queueName := "salas" // Queue de RabbitMQ se mantiene igual
	// Key única: incluye puntero de la conexión para que cada dispositivo sea único
	connectionKey := fmt.Sprintf("general_%s_%p", userId, c)

	log.Println("🛰️ Usuario conectado (general):", userId)

	wsUsuariosManagers.addConnection(connectionKey, c)

	cm := newConnectionManager(func() {
		log.Println("🧹 Iniciando cleanup para usuario (general):", userId)
		wsUsuariosManagers.removeConnection(connectionKey, c)

		if err := c.Close(); err != nil {
			log.Printf("⚠️ Error al cerrar conexión WS: %v", err)
		}

		log.Println("❌ Cliente desconectado y limpieza completada (general):", userId)
	})
	defer cm.close()

	s.startQueueConsumer(queueName, managerTypeUsuarios, connectionKey, cm)
	s.readLoop(c, cm, fmt.Sprintf("usuario %s (general)", userId))
}

// ---------- USO POR SALA ----------
func (s SalaHandlerWS) UsoSala(c *websocket.Conn) {
	userId := fmt.Sprintf("%v", c.Locals("userId"))
	salaId := c.Params("salaId")
	queueName := "salas_" + salaId
	// Key única: incluye puntero de la conexión para que cada dispositivo sea único
	connectionKey := fmt.Sprintf("sala_%s_%s_%p", salaId, userId, c)

	log.Println("🛰️ Usuario conectado:", userId, "a sala:", salaId)

	wsUsuariosBySalaManagers.addConnection(connectionKey, c)

	cm := newConnectionManager(func() {
		log.Println("🧹 Iniciando cleanup para sala:", salaId)
		wsUsuariosBySalaManagers.removeConnection(connectionKey, c)

		if err := c.Close(); err != nil {
			log.Printf("⚠️ Error al cerrar conexión WS: %v", err)
		}

		log.Println("❌ Cliente desconectado y limpieza completada - Usuario:", userId, "Sala:", salaId)
	})
	defer cm.close()

	s.startQueueConsumer(queueName, managerTypeSala, connectionKey, cm)
	s.readLoop(c, cm, fmt.Sprintf("usuario %s (sala %s)", userId, salaId))
}

// ---------- CONSTRUCTOR ----------
func NewSalaHandlerWS(salaService port.SalaService, rabbitMQService port.RabbitMQService) *SalaHandlerWS {
	return &SalaHandlerWS{salaService: salaService, rabbitMQService: rabbitMQService}
}

var _ port.SalaHandlerWS = (*SalaHandlerWS)(nil)
