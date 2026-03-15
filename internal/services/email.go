package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"strings"

	"github.com/spf13/viper"

	"wishlist/internal/config"
)

type EmailServiceImpl struct {
	host     string
	port     string
	user     string
	password string
	from     string
	domain   string
	sender   func(to, subject, body string) error
}

func NewEmailService() *EmailServiceImpl {
	es := &EmailServiceImpl{
		host:     viper.GetString(config.EmailHost),
		port:     viper.GetString(config.EmailPort),
		user:     viper.GetString(config.EmailUser),
		password: viper.GetString(config.EmailPassword),
		from:     viper.GetString(config.EmailFrom),
		domain:   getDomainURL(viper.GetString(config.WebAppDomain)),
	}
	es.sender = es.send
	return es
}

func getDomainURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	var host string
	var path string

	if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
		host = parsed.Host
		path = strings.TrimSuffix(parsed.EscapedPath(), "/")
	} else {
		host = strings.TrimPrefix(strings.TrimPrefix(value, "http://"), "https://")
		host = strings.TrimSuffix(host, "/")
	}

	scheme := "https"
	lowerHost := strings.ToLower(host)
	if strings.HasPrefix(lowerHost, "localhost") || strings.HasPrefix(lowerHost, "127.0.0.1") || strings.HasPrefix(lowerHost, "[::1]") {
		scheme = "http"
	}

	return strings.TrimSuffix(fmt.Sprintf("%s://%s%s", scheme, host, path), "/")
}

func (svc *EmailServiceImpl) SendEmailVerificationLetter(_ context.Context, to, token string) error {
	body := fmt.Sprintf("Please verify your email address by clicking the link below:\n\n"+
		"%s",
		fmt.Sprintf("%s/verify-email?token=%s", svc.domain, token))
	return svc.sendEmail(to, "Verify your email", body)
}

func (svc *EmailServiceImpl) SendPasswordResetLetter(_ context.Context, to, token string) error {
	body := fmt.Sprintf("You requested a password reset.\n\n"+
		"Click the link below to set a new password:\n\n"+
		"%s\n\n"+
		"The link expires in 1 hour. If you didn't request this, ignore this email.",
		fmt.Sprintf("%s/reset-password?token=%s", svc.domain, token))
	return svc.sendEmail(to, "Reset your password", body)
}

func buildMessage(from, to, subject, body string) []byte {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return []byte(sb.String())
}

func (svc *EmailServiceImpl) sendEmail(to, subject, body string) error {
	if svc.sender != nil {
		return svc.sender(to, subject, body)
	}
	return svc.send(to, subject, body)
}

func (svc *EmailServiceImpl) send(to, subject, body string) error {
	msg := buildMessage(svc.from, to, subject, body)
	addr := net.JoinHostPort(svc.host, svc.port)
	auth := smtp.PlainAuth("", svc.user, svc.password, svc.host)

	if svc.port == "465" {
		return svc.sendTLS(addr, auth, to, msg)
	}

	return smtp.SendMail(addr, auth, svc.user, []string{to}, msg)
}

func (svc *EmailServiceImpl) sendTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: svc.host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, svc.host)
	if err != nil {
		return fmt.Errorf("creating smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authorizing smtp client: %w", err)
	}

	if err = client.Mail(svc.user); err != nil {
		return fmt.Errorf("smtp: issuing mail command: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp: issuing recipient command: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: issuing data command: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp: writing: %w", err)
	}

	return w.Close()
}
