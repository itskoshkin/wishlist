package events

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/spf13/viper"

	"wishlist/internal/config"
)

const (
	emailTopic = "notifications.email"
)

type Type string

const (
	TypeEmailVerification Type = "email.verification"
	TypePasswordReset     Type = "email.password_reset"
)

type Envelope struct {
	ID        string          `json:"id"`
	Type      Type            `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type EmailVerificationPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

type PasswordResetPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func EmailTopic() string {
	prefix := strings.Trim(viper.GetString(config.KafkaTopicPrefix), ". ")
	if prefix == "" {
		return emailTopic
	}

	return prefix + "." + emailTopic
}
