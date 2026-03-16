package events

import "context"

type EmailSender struct {
	publisher *Publisher
}

func NewEmailSender(publisher *Publisher) *EmailSender {
	return &EmailSender{publisher: publisher}
}

func (s *EmailSender) SendPasswordReset(ctx context.Context, userID, to, token string) error {
	return s.publisher.PublishPasswordReset(ctx, PasswordResetPayload{
		UserID: userID,
		Email:  to,
		Token:  token,
	})
}

func (s *EmailSender) SendEmailVerification(ctx context.Context, userID, to, token string) error {
	return s.publisher.PublishEmailVerification(ctx, EmailVerificationPayload{
		UserID: userID,
		Email:  to,
		Token:  token,
	})
}
