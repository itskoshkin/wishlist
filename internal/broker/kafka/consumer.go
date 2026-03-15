package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"

	"wishlist/internal/broker"
	"wishlist/internal/logger"
)

type ConsumerConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupID string   `yaml:"group_id"`
	Auth    AuthConfig
}

type Consumer struct {
	cfg    ConsumerConfig
	reader *kafka.Reader
	dialer *kafka.Dialer
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	dialer, err := newDialer(cfg.Auth)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		cfg:    cfg,
		dialer: dialer,
	}, nil
}

func (c *Consumer) Subscribe(ctx context.Context, topic string, handler broker.Handler) error {
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.cfg.Brokers,
		Topic:   topic,
		GroupID: c.cfg.GroupID,
		Dialer:  c.dialer,
	})

	logger.Info("Subscribed to topic '%s'.", topic)

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // graceful shutdown
			}
			return err
		}

		if err = handler(ctx, msg.Value); err != nil {
			logger.Error("Failed to handle message from topic '%s' at offset %d: %v", topic, msg.Offset, err)
			continue
		}

		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			logger.Error("Failed to commit message from topic '%s': %v", topic, err)
		}
	}
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}

	return nil
}
