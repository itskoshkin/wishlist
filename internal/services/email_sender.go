package services

import "context"

type SMTPEmailSender struct {
	email EmailService
}

func NewSMTPEmailSender(email EmailService) *SMTPEmailSender {
	return &SMTPEmailSender{email: email}
}

func (s *SMTPEmailSender) SendPasswordReset(ctx context.Context, _ string, to, token string) error {
	return s.email.SendPasswordResetLetter(ctx, to, token)
}

func (s *SMTPEmailSender) SendEmailVerification(ctx context.Context, _ string, to, token string) error {
	return s.email.SendEmailVerificationLetter(ctx, to, token)
}
