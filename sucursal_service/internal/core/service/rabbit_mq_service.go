package service

import (
	"encoding/json"
	"log"
	"multiroom/sucursal-service/internal/core/port"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type consumerEntry struct {
	queueName string
	handler   func(msg amqp.Delivery)
	args      amqp.Table
}

type RabbitMQService struct {
	conn              *amqp.Connection
	publishCh         *amqp.Channel
	url               string
	mutex             sync.Mutex
	connected         bool
	consumers         []consumerEntry
	declaredQueues    map[string]bool
	declaredExchanges map[string]bool
}

func NewRabbitMQService(url string) *RabbitMQService {
	s := &RabbitMQService{
		url:               url,
		declaredQueues:    make(map[string]bool),
		declaredExchanges: make(map[string]bool),
	}
	go s.reconnectLoop()
	return s
}

func (r *RabbitMQService) reconnectLoop() {
	for {
		if !r.connected {
			log.Println("🔌 Intentando conectar a RabbitMQ...")
			conn, err := amqp.Dial(r.url)
			if err != nil {
				log.Printf("❌ Error conectando a RabbitMQ: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			ch, err := conn.Channel()
			if err != nil {
				log.Printf("❌ Error creando canal de publicación: %v", err)
				_ = conn.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			r.mutex.Lock()
			r.conn = conn
			r.publishCh = ch
			r.connected = true
			r.mutex.Unlock()

			log.Println("✅ Conectado a RabbitMQ")

			// Escuchar cierres inesperados
			go func() {
				err := <-conn.NotifyClose(make(chan *amqp.Error))
				log.Printf("⚠️ Conexión cerrada: %v", err)
				r.mutex.Lock()
				r.connected = false
				if r.publishCh != nil {
					_ = r.publishCh.Close()
					r.publishCh = nil
				}
				r.conn = nil
				r.declaredQueues = make(map[string]bool)
				r.declaredExchanges = make(map[string]bool)
				r.mutex.Unlock()
			}()

			// Reiniciar consumidores
			r.mutex.Lock()
			consumers := make([]consumerEntry, len(r.consumers))
			copy(consumers, r.consumers)
			r.mutex.Unlock()

			for _, entry := range consumers {
				go func(e consumerEntry) {
					err := r.StartConsumer(e.queueName, e.handler, e.args)
					if err != nil {
						log.Printf("❌ Error reiniciando consumidor %s: %v", e.queueName, err)
					}
				}(entry)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (r *RabbitMQService) GetChannel() (*amqp.Channel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected || r.conn == nil {
		return nil, amqp.ErrClosed
	}
	return r.conn.Channel()
}

func (r *RabbitMQService) PublishToExchange(exchange string, body interface{}) error {
	r.mutex.Lock()
	ch := r.publishCh
	declared := r.declaredExchanges[exchange]
	r.mutex.Unlock()

	if ch == nil {
		return amqp.ErrClosed
	}

	if !declared {
		err := ch.ExchangeDeclare(
			exchange,
			"fanout",
			false, // durable
			false, // autoDelete
			false, // internal
			false, // noWait
			nil,
		)
		if err != nil {
			log.Printf("Error declarando exchange %s: %v", exchange, err)
			return err
		}
		r.mutex.Lock()
		r.declaredExchanges[exchange] = true
		r.mutex.Unlock()
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error convirtiendo body a JSON: %v", err)
		return err
	}

	err = ch.Publish(
		exchange,
		"", // routing key vacío para fanout
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonBody,
			DeliveryMode: amqp.Transient,
		},
	)
	if err != nil {
		log.Printf("Error publicando en exchange %s: %v", exchange, err)
		return err
	}

	return nil
}

func (r *RabbitMQService) Publish(queueName string, body interface{}, args amqp.Table) error {
	r.mutex.Lock()
	conn := r.conn
	connected := r.connected
	r.mutex.Unlock()

	if !connected || conn == nil {
		return amqp.ErrClosed
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer func() {
		_ = ch.Close()

	}()

	// Declarar cola si no existe
	_, err = ch.QueueDeclare(
		queueName,
		false,
		true,
		false,
		false,
		args,
	)
	if err != nil {
		log.Printf("Error declarando cola %s: %v", queueName, err)
		return err
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonBody,
			DeliveryMode: amqp.Transient,
		},
	)
	if err != nil {
		log.Printf("Error publicando mensaje en cola %s: %v", queueName, err)
		return err
	}

	return nil
}

func (r *RabbitMQService) StartConsumer(queueName string, handler func(amqp.Delivery), args amqp.Table) error {
	go func() {
		r.mutex.Lock()
		alreadyRegistered := false
		for _, c := range r.consumers {
			if c.queueName == queueName {
				alreadyRegistered = true
				break
			}
		}
		if !alreadyRegistered {
			r.consumers = append(r.consumers, consumerEntry{
				queueName: queueName,
				handler:   handler,
				args:      args,
			})
		}
		r.mutex.Unlock()

		for {
			r.mutex.Lock()
			conn := r.conn
			connected := r.connected
			r.mutex.Unlock()

			if !connected || conn == nil {
				time.Sleep(2 * time.Second)
				continue
			}

			channel, err := conn.Channel()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			q, err := channel.QueueDeclare(
				queueName,
				false,
				true, // autoDelete
				false,
				false,
				args,
			)
			if err != nil {
				_ = channel.Close()
				time.Sleep(2 * time.Second)
				continue
			}

			msgs, err := channel.Consume(
				q.Name,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				_ = channel.Close()
				time.Sleep(2 * time.Second)
				continue
			}

			for msg := range msgs {
				handler(msg)
			}

			_ = channel.Close()
			time.Sleep(2 * time.Second)
		}
	}()
	return nil
}

var _ port.RabbitMQService = (*RabbitMQService)(nil)
