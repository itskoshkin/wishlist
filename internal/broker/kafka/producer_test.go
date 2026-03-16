package kafka

import "testing"

func TestNewSASLMechanismPlain(t *testing.T) {
	mechanism, err := newSASLMechanism(AuthConfig{
		Enabled:   true,
		Mechanism: "plain",
		Username:  "wishlist",
		Password:  "secret",
	})
	if err != nil {
		t.Fatalf("newSASLMechanism() error = %v", err)
	}
	if mechanism == nil {
		t.Fatal("newSASLMechanism() returned nil mechanism")
	}
	if got := mechanism.Name(); got != "PLAIN" {
		t.Fatalf("mechanism.Name() = %q, want %q", got, "PLAIN")
	}
}

func TestNewSASLMechanismRejectsUnsupportedMechanism(t *testing.T) {
	_, err := newSASLMechanism(AuthConfig{
		Enabled:   true,
		Mechanism: "scram-sha-256",
		Username:  "wishlist",
		Password:  "secret",
	})
	if err == nil {
		t.Fatal("newSASLMechanism() error = nil, want unsupported mechanism error")
	}
}
