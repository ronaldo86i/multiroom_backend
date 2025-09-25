package port

import amqp "github.com/rabbitmq/amqp091-go"

type RabbitMQService interface {
	//GetChannel() (*amqp.Channel, error)
	StartConsumer(queueName string, handler func(amqp.Delivery), args amqp.Table) error
	PublishToExchange(exchange string, body interface{}) error
	Publish(queueName string, body interface{}, args amqp.Table) error
}
