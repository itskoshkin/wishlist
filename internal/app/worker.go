package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"

	"wishlist/internal/broker"
	"wishlist/internal/broker/kafka"
	"wishlist/internal/broker/rabbitmq"
	"wishlist/internal/config"
	"wishlist/internal/events"
	"wishlist/internal/logger"
	"wishlist/internal/services"
)

type Worker struct {
	consumer broker.Consumer
	emailSvc *services.EmailServiceImpl
}

func LoadWorker() *Worker {
	config.LoadConfig()
	logger.SetupLogger()

	consumer, err := newBrokerConsumer()
	if err != nil {
		logger.Fatal(err)
	}

	return &Worker{
		consumer: consumer,
		emailSvc: services.NewEmailService(),
	}
}

func (w *Worker) Run() {
	defer w.closeConsumer()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	topic := events.EmailTopic()
	logger.Info("Worker is listening to '%s' via %s.", topic, config.CurrentBrokerType())

	if err := w.consumer.Subscribe(ctx, topic, w.handleEmailEvent); err != nil {
		logger.Fatalf("Worker stopped unexpectedly: %v", err)
	}

	logger.Info("Worker stopped.")
}

func newBrokerConsumer() (broker.Consumer, error) {
	switch config.CurrentBrokerType() {
	case "", "none":
		return nil, fmt.Errorf("broker is disabled")
	case "kafka":
		return kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: viper.GetStringSlice(config.KafkaBrokers),
			GroupID: viper.GetString(config.KafkaGroupID),
			Auth: kafka.AuthConfig{
				Enabled:   viper.GetBool(config.KafkaAuthEnabled),
				Mechanism: viper.GetString(config.KafkaAuthMechanism),
				Username:  viper.GetString(config.KafkaAuthUsername),
				Password:  viper.GetString(config.KafkaAuthPassword),
				UseTLS:    viper.GetBool(config.KafkaAuthTLS),
			},
		})
	case "rabbitmq":
		return rabbitmq.NewConsumer(rabbitmq.ConsumerConfig{
			URL: viper.GetString(config.RabbitMQURL),
		})
	default:
		return nil, fmt.Errorf("unsupported broker type for worker: %s", config.CurrentBrokerType())
	}
}

func (w *Worker) closeConsumer() {
	if w.consumer == nil {
		return
	}
	if err := w.consumer.Close(); err != nil {
		logger.Error("Failed to close broker consumer: %v", err)
	}
}

func (w *Worker) handleEmailEvent(ctx context.Context, msg []byte) error {
	var env events.Envelope
	if err := json.Unmarshal(msg, &env); err != nil {
		return fmt.Errorf("unmarshal event envelope: %w", err)
	}

	switch env.Type {
	case events.TypeEmailVerification:
		var payload events.EmailVerificationPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal verification payload: %w", err)
		}

		return w.emailSvc.SendEmailVerificationLetter(ctx, payload.Email, payload.Token)

	case events.TypePasswordReset:
		var payload events.PasswordResetPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal password reset payload: %w", err)
		}

		return w.emailSvc.SendPasswordResetLetter(ctx, payload.Email, payload.Token)

	default:
		return fmt.Errorf("unsupported event type: %s", env.Type)
	}
}
