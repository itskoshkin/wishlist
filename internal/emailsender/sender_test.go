package emailsender

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wishlist/internal/events"
)

type emailServiceMock struct {
	verificationCalls int
	resetCalls        int
	lastTo            string
	lastToken         string
}

func (m *emailServiceMock) SendPasswordResetLetter(_ context.Context, to, token string) error {
	m.resetCalls++
	m.lastTo = to
	m.lastToken = token
	return nil
}

func (m *emailServiceMock) SendEmailVerificationLetter(_ context.Context, to, token string) error {
	m.verificationCalls++
	m.lastTo = to
	m.lastToken = token
	return nil
}

func TestSender_HandleEmailEvent_Verification(t *testing.T) {
	emailSvc := &emailServiceMock{}
	sender := &Sender{emailSvc: emailSvc}

	msg := mustMarshalEvent(t, events.TypeEmailVerification, events.EmailVerificationPayload{
		UserID: "user-1",
		Email:  "alice@example.com",
		Token:  "verify-token",
	})

	if err := sender.handleEmailEvent(context.Background(), msg); err != nil {
		t.Fatalf("handleEmailEvent() error = %v", err)
	}
	if emailSvc.verificationCalls != 1 {
		t.Fatalf("verificationCalls = %d, want 1", emailSvc.verificationCalls)
	}
	if emailSvc.resetCalls != 0 {
		t.Fatalf("resetCalls = %d, want 0", emailSvc.resetCalls)
	}
	if emailSvc.lastTo != "alice@example.com" {
		t.Fatalf("lastTo = %q, want %q", emailSvc.lastTo, "alice@example.com")
	}
	if emailSvc.lastToken != "verify-token" {
		t.Fatalf("lastToken = %q, want %q", emailSvc.lastToken, "verify-token")
	}
}

func TestSender_HandleEmailEvent_PasswordReset(t *testing.T) {
	emailSvc := &emailServiceMock{}
	sender := &Sender{emailSvc: emailSvc}

	msg := mustMarshalEvent(t, events.TypePasswordReset, events.PasswordResetPayload{
		UserID: "user-2",
		Email:  "bob@example.com",
		Token:  "reset-token",
	})

	if err := sender.handleEmailEvent(context.Background(), msg); err != nil {
		t.Fatalf("handleEmailEvent() error = %v", err)
	}
	if emailSvc.resetCalls != 1 {
		t.Fatalf("resetCalls = %d, want 1", emailSvc.resetCalls)
	}
	if emailSvc.verificationCalls != 0 {
		t.Fatalf("verificationCalls = %d, want 0", emailSvc.verificationCalls)
	}
	if emailSvc.lastTo != "bob@example.com" {
		t.Fatalf("lastTo = %q, want %q", emailSvc.lastTo, "bob@example.com")
	}
	if emailSvc.lastToken != "reset-token" {
		t.Fatalf("lastToken = %q, want %q", emailSvc.lastToken, "reset-token")
	}
}

func mustMarshalEvent(t *testing.T, eventType events.Type, payload any) []byte {
	t.Helper()

	payloadData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(payload) error = %v", err)
	}

	msg, err := json.Marshal(events.Envelope{
		ID:        "event-1",
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payloadData,
	})
	if err != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", err)
	}

	return msg
}
