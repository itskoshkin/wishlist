package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type ProducerConfig struct {
	URL string `yaml:"url"`
}

type Producer struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq open channel: %w", err)
	}

	return &Producer{
		conn: conn,
		ch:   ch,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, msg []byte) error {
	q, err := p.ch.QueueDeclare(
		topic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", topic, err)
	}

	return p.ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "application/json",
			Body:         msg,
		},
	)
}

func (p *Producer) Close() error {
	if err := p.ch.Close(); err != nil {
		_ = p.conn.Close()
		return fmt.Errorf("close channel: %w", err)
	}

	return p.conn.Close()
}
