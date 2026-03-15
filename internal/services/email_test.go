package services

import (
	"context"
	"net/smtp"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"wishlist/internal/config"
)

func setEmailConfigForTests() {
	viper.Reset()
	viper.Set(config.EmailHost, "smtp.example.com")
	viper.Set(config.EmailPort, "587")
	viper.Set(config.EmailUser, "noreply@example.com")
	viper.Set(config.EmailPassword, "secret")
	viper.Set(config.EmailFrom, "Wishlist <noreply@example.com>")
	viper.Set(config.WebAppDomain, "wishlist.example.com")
}

func TestNewEmailService_UsesConfig(t *testing.T) {
	setEmailConfigForTests()

	svc := NewEmailService()

	if svc.host != "smtp.example.com" {
		t.Fatalf("host = %s, want smtp.example.com", svc.host)
	}
	if svc.port != "587" {
		t.Fatalf("port = %s, want 587", svc.port)
	}
	if svc.user != "noreply@example.com" {
		t.Fatalf("user = %s, want noreply@example.com", svc.user)
	}
	if svc.from != "Wishlist <noreply@example.com>" {
		t.Fatalf("from = %s, want Wishlist <noreply@example.com>", svc.from)
	}
	if svc.domain != "https://wishlist.example.com" {
		t.Fatalf("domain = %s, want https://wishlist.example.com", svc.domain)
	}
	if svc.sender == nil {
		t.Fatal("sender is nil, want initialized sender")
	}
}

func TestNewEmailService_UsesHTTPForLocalhostDomain(t *testing.T) {
	setEmailConfigForTests()
	viper.Set(config.WebAppDomain, "localhost:8080")

	svc := NewEmailService()

	if svc.domain != "http://localhost:8080" {
		t.Fatalf("domain = %s, want http://localhost:8080", svc.domain)
	}
}

func TestBuildMessage(t *testing.T) {
	msg := string(buildMessage("from@example.com", "to@example.com", "Hello", "Line 1\nLine 2"))

	if !strings.Contains(msg, "From: from@example.com\r\n") {
		t.Fatal("message does not contain From header")
	}
	if !strings.Contains(msg, "To: to@example.com\r\n") {
		t.Fatal("message does not contain To header")
	}
	if !strings.Contains(msg, "Subject: Hello\r\n") {
		t.Fatal("message does not contain Subject header")
	}
	if !strings.Contains(msg, "Content-Type: text/plain; charset=UTF-8\r\n") {
		t.Fatal("message does not contain content type header")
	}
	if !strings.HasSuffix(msg, "Line 1\nLine 2") {
		t.Fatal("message body mismatch")
	}
}

func TestEmailService_SendEmailVerificationLetter_ComposesVerificationLink(t *testing.T) {
	svc := &EmailServiceImpl{domain: "https://wishlist.example.com"}

	var gotTo, gotSubject, gotBody string
	svc.sender = func(to, subject, body string) error {
		gotTo = to
		gotSubject = subject
		gotBody = body
		return nil
	}

	err := svc.SendEmailVerificationLetter(context.Background(), "alice@example.com", "token-123")
	if err != nil {
		t.Fatalf("SendEmailVerificationLetter() error = %v", err)
	}

	if gotTo != "alice@example.com" {
		t.Fatalf("to = %s, want alice@example.com", gotTo)
	}
	if gotSubject != "Verify your email" {
		t.Fatalf("subject = %s, want Verify your email", gotSubject)
	}
	if !strings.Contains(gotBody, "https://wishlist.example.com/verify-email?token=token-123") {
		t.Fatalf("verification body does not contain expected link: %s", gotBody)
	}
}

func TestEmailService_SendPasswordResetLetter_ComposesResetLink(t *testing.T) {
	svc := &EmailServiceImpl{domain: "https://wishlist.example.com"}

	var gotTo, gotSubject, gotBody string
	svc.sender = func(to, subject, body string) error {
		gotTo = to
		gotSubject = subject
		gotBody = body
		return nil
	}

	err := svc.SendPasswordResetLetter(context.Background(), "alice@example.com", "reset-456")
	if err != nil {
		t.Fatalf("SendPasswordResetLetter() error = %v", err)
	}

	if gotTo != "alice@example.com" {
		t.Fatalf("to = %s, want alice@example.com", gotTo)
	}
	if gotSubject != "Reset your password" {
		t.Fatalf("subject = %s, want Reset your password", gotSubject)
	}
	if !strings.Contains(gotBody, "https://wishlist.example.com/reset-password?token=reset-456") {
		t.Fatalf("reset body does not contain expected link: %s", gotBody)
	}
	if !strings.Contains(gotBody, "The link expires in 1 hour") {
		t.Fatalf("reset body missing expiration note: %s", gotBody)
	}
}

func TestEmailService_send_SMTPError(t *testing.T) {
	svc := &EmailServiceImpl{
		host:     "127.0.0.1",
		port:     "1",
		user:     "noreply@example.com",
		password: "secret",
		from:     "Wishlist <noreply@example.com>",
	}

	err := svc.send("alice@example.com", "Subject", "Body")
	if err == nil {
		t.Fatal("send() error = nil, want smtp error")
	}
}

func TestEmailService_sendTLS_DialError(t *testing.T) {
	svc := &EmailServiceImpl{host: "127.0.0.1"}
	auth := smtp.PlainAuth("", "user", "pass", "127.0.0.1")
	err := svc.sendTLS("127.0.0.1:1", auth, "alice@example.com", []byte("msg"))
	if err == nil {
		t.Fatal("sendTLS() error = nil, want dial error")
	}
	if !strings.Contains(err.Error(), "tls dial") {
		t.Fatalf("sendTLS() error = %v, want tls dial prefix", err)
	}
}
