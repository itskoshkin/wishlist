package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"wishlist/internal/broker"
	"wishlist/internal/logger"
)

type ConsumerConfig struct {
	URL string `yaml:"url"`
}

type Consumer struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq open channel: %w", err)
	}

	if err = ch.Qos(1, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq set qos: %w", err)
	}

	return &Consumer{
		conn: conn,
		ch:   ch,
	}, nil
}

func (c *Consumer) Subscribe(ctx context.Context, topic string, handler broker.Handler) error {
	q, err := c.ch.QueueDeclare(
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

	messages, err := c.ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume %s: %w", topic, err)
	}

	logger.Info("Subscribed to queue '%s'.", topic)

	for {
		select {
		case <-ctx.Done():
			return nil

		case d, ok := <-messages:
			if !ok {
				return fmt.Errorf("channel closed for queue %s", topic)
			}

			if err = handler(ctx, d.Body); err != nil {
				logger.Error("Failed to handle message from queue '%s': %v", topic, err)
				_ = d.Nack(false, true)
				continue
			}

			_ = d.Ack(false)
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.ch.Close(); err != nil {
		_ = c.conn.Close()
		return fmt.Errorf("close channel: %w", err)
	}

	return c.conn.Close()
}
