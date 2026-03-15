package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/broker"
	"wishlist/internal/broker/kafka"
	"wishlist/internal/broker/rabbitmq"
	"wishlist/internal/config"
)

func NewEventPublisher() (*Publisher, error) {
	switch config.CurrentBrokerType() {
	case "", "none":
		return nil, nil
	case "kafka":
		producer, err := kafka.NewProducer(kafka.ProducerConfig{
			Brokers: viper.GetStringSlice(config.KafkaBrokers),
			Auth: kafka.AuthConfig{
				Enabled:   viper.GetBool(config.KafkaAuthEnabled),
				Mechanism: viper.GetString(config.KafkaAuthMechanism),
				Username:  viper.GetString(config.KafkaAuthUsername),
				Password:  viper.GetString(config.KafkaAuthPassword),
				UseTLS:    viper.GetBool(config.KafkaAuthTLS),
			},
		})
		if err != nil {
			return nil, err
		}
		return NewPublisher(producer), nil
	case "rabbitmq":
		producer, err := rabbitmq.NewProducer(rabbitmq.ProducerConfig{
			URL: viper.GetString(config.RabbitMQURL),
		})
		if err != nil {
			return nil, err
		}
		return NewPublisher(producer), nil
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", config.CurrentBrokerType())
	}
}

type Publisher struct {
	producer broker.Producer
}

func NewPublisher(producer broker.Producer) *Publisher {
	return &Publisher{producer: producer}
}

func (p *Publisher) PublishEmailVerification(ctx context.Context, payload EmailVerificationPayload) error {
	return p.publish(ctx, EmailTopic(), TypeEmailVerification, payload)
}

func (p *Publisher) PublishPasswordReset(ctx context.Context, payload PasswordResetPayload) error {
	return p.publish(ctx, EmailTopic(), TypePasswordReset, payload)
}

func (p *Publisher) publish(ctx context.Context, topic string, eventType Type, payload any) error {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	env := Envelope{
		ID:        uuid.NewString(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payloadData,
	}

	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.producer.Publish(ctx, topic, data)
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}
