package internal

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

func NewClient(config *Config) (*RabbitClient, error) {
	conn, err := amqp.Dial(config.EventBus.uri())
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitClient{
		conn,
		ch,
	}, nil
}

func (c *RabbitClient) Close() error {
	_ = c.ch.Close()
	return c.conn.Close()
}

func (c *RabbitClient) CreateQueue(routingKey string) error {
	// Declaring a queue is idempotent
	_, err := c.ch.QueueDeclare(
		routingKey,
		RabbitDurable,
		RabbitAutoDelete,
		RabbitExclusive,
		RabbitNoWait,
		nil,
	)

	return err
}

func (c *RabbitClient) Publish(
	routingKey string,
	eventType EventType,
	msg interface{},
) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.ch.Publish(
		RabbitExchange,
		routingKey,
		RabbitMandatory,
		RabbitImmediate,
		amqp.Publishing{
			Type:        string(eventType),
			ContentType: RabbitContentType,
			Body:        body,
		},
	)
}

func (c *RabbitClient) Listen(queueName string) (<-chan amqp.Delivery, error) {
	return c.ch.Consume(
		queueName,
		RabbitConsumer,
		RabbitAutoAck,
		RabbitExclusive,
		RabbitNoLocal,
		RabbitNoWait,
		nil,
	)
}
