package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type ProducerConfig struct {
	Brokers []string `yaml:"brokers"`
	Auth    AuthConfig
}

type AuthConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Mechanism string `yaml:"mechanism"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	UseTLS    bool   `yaml:"tls"`
}

type Producer struct {
	writers map[string]*kafka.Writer
	cfg     ProducerConfig
	dialer  *kafka.Dialer
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	dialer, err := newDialer(cfg.Auth)
	if err != nil {
		return nil, err
	}

	return &Producer{
		writers: make(map[string]*kafka.Writer),
		cfg:     cfg,
		dialer:  dialer,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, msg []byte) error {
	w, ok := p.writers[topic]
	if !ok {
		w = kafka.NewWriter(kafka.WriterConfig{
			Brokers:      p.cfg.Brokers,
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: int(kafka.RequireOne),
			Dialer:       p.dialer,
		})
		p.writers[topic] = w
	}

	return w.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (p *Producer) Close() error {
	var errs []error

	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close writer %s: %w", topic, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close producers: %v", errs)
	}

	return nil
}

func newDialer(cfg AuthConfig) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{}

	if cfg.UseTLS {
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	if !cfg.Enabled {
		return dialer, nil
	}

	mechanism, err := newSASLMechanism(cfg)
	if err != nil {
		return nil, err
	}

	dialer.SASLMechanism = mechanism

	return dialer, nil
}

func newSASLMechanism(cfg AuthConfig) (sasl.Mechanism, error) {
	username := strings.TrimSpace(cfg.Username)
	password := strings.TrimSpace(cfg.Password)
	mechanism := strings.ToLower(strings.TrimSpace(cfg.Mechanism))
	if mechanism == "" {
		mechanism = "plain"
	}

	if username == "" {
		return nil, fmt.Errorf("kafka auth username is required")
	}
	if password == "" {
		return nil, fmt.Errorf("kafka auth password is required")
	}

	switch mechanism {
	case "plain":
		return plain.Mechanism{
			Username: username,
			Password: password,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported kafka sasl mechanism: %s", cfg.Mechanism)
	}
}
