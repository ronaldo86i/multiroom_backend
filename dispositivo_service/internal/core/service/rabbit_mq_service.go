package service

import (
	"encoding/json"
	"fmt"
	"log"
	"multiroom/dispositivo-service/internal/core/port"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type consumerEntry struct {
	queueName string
	handler   func(msg amqp.Delivery)
}

type RabbitMQService struct {
	conn      *amqp.Connection
	url       string
	mutex     sync.Mutex
	connected bool
	consumers []consumerEntry
}

func NewRabbitMQService(url string) *RabbitMQService {
	s := &RabbitMQService{
		url: url,
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

			r.mutex.Lock()
			r.conn = conn
			r.connected = true
			r.mutex.Unlock()
			log.Println("✅ Conectado a RabbitMQ")

			// Escuchar cierres inesperados
			go func() {
				err := <-conn.NotifyClose(make(chan *amqp.Error))
				log.Printf("⚠️ Conexión cerrada: %v", err)
				r.mutex.Lock()
				r.connected = false
				r.conn = nil
				r.mutex.Unlock()
			}()

			// Reiniciar consumidores
			r.mutex.Lock()
			consumers := make([]consumerEntry, len(r.consumers))
			copy(consumers, r.consumers)
			r.mutex.Unlock()

			for _, entry := range consumers {
				go func(e consumerEntry) {
					err := r.StartConsumer(e.queueName, e.handler)
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
	channel, err := r.GetChannel()
	if err != nil {
		log.Printf("❌ No se pudo obtener canal para publicar: %v", err)
		return err
	}
	defer func() {
		_ = channel.Close()
	}()

	err = channel.ExchangeDeclare(
		exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("❌ Error declarando exchange: %v", err)
		return err
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("❌ Error convirtiendo body a JSON: %v", err)
		return err
	}

	err = channel.Publish(
		exchange,
		"", // routing key vacío para fanout
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonBody,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		log.Printf("❌ Error publicando mensaje en RabbitMQ: %v", err)
		return err
	}
	log.Printf("✅ Mensaje publicado en exchange %s", exchange)
	return nil
}

func (r *RabbitMQService) Publish(queueName string, body interface{}) error {
	channel, err := r.GetChannel()
	if err != nil {
		log.Printf("❌ No se pudo obtener canal para publicar: %v", err)
		return err
	}
	defer func() { _ = channel.Close() }()

	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Printf("❌ Error declarando cola para publicar: %v", err)
		return err
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("❌ Error convirtiendo body a JSON: %v", err)
		return err
	}

	err = channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonBody,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		log.Printf("❌ Error publicando mensaje en RabbitMQ: %v", err)
		return err
	}

	log.Printf("✅ Mensaje publicado en cola %s", queueName)
	return nil
}

func (r *RabbitMQService) StartConsumer(queueName string, handler func(amqp.Delivery)) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("error creando canal: %w", err)
	}

	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error declarando cola: %w", err)
	}

	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil) // autoAck: false
	if err != nil {
		return fmt.Errorf("error creando consumidor: %w", err)
	}

	go func() {
		for msg := range msgs {
			handler(msg)
		}
	}()

	return nil
}

var _ port.RabbitMQService = (*RabbitMQService)(nil)
