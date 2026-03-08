package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

type tokenStorageMock struct {
	checkRevoked bool
	checkErr     error

	revokeErr   error
	revokeCalls int
	revokeIDs   []string
	revokeTTLs  []time.Duration
}

func (m *tokenStorageMock) SaveEmailVerificationToken(ctx context.Context, tokenID, userID string) error {
	return nil
}

func (m *tokenStorageMock) GetEmailVerificationToken(ctx context.Context, tokenID string) (string, error) {
	return "", nil
}

func (m *tokenStorageMock) DeleteEmailVerificationToken(ctx context.Context, tokenID string) error {
	return nil
}

func (m *tokenStorageMock) CheckIfAuthTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	return m.checkRevoked, m.checkErr
}

func (m *tokenStorageMock) RevokeAuthTokens(ctx context.Context, tokenID string, remainingTTL time.Duration) error {
	m.revokeCalls++
	m.revokeIDs = append(m.revokeIDs, tokenID)
	m.revokeTTLs = append(m.revokeTTLs, remainingTTL)
	return m.revokeErr
}

func (m *tokenStorageMock) SavePasswordResetToken(ctx context.Context, tokenID string, userID string) error {
	return nil
}

func (m *tokenStorageMock) GetPasswordResetToken(ctx context.Context, tokenID string) (string, error) {
	return "", nil
}

func (m *tokenStorageMock) DeletePasswordResetToken(ctx context.Context, tokenID string) error {
	return nil
}

func setAuthConfigForTests() {
	viper.Reset()
	viper.Set(config.AccessTokenSecret, "access-secret-for-tests")
	viper.Set(config.RefreshTokenSecret, "refresh-secret-for-tests")
	viper.Set(config.AccessTokenTTL, "1h")
	viper.Set(config.RefreshTokenTTL, "2h")
	viper.Set(config.JwtIssuer, "wishlist-test")
	viper.Set(config.JwtAudience, []string{"wishlist-api-test"})
}

func TestAuthService_GenerateAndValidateTokens(t *testing.T) {
	setAuthConfigForTests()
	storage := &tokenStorageMock{}
	svc := NewAuthService(storage)

	userID := uuid.New()
	accessToken, refreshToken, err := svc.GenerateTokens(userID)
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}

	accessUserID, err := svc.ValidateAccessToken(context.Background(), accessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if accessUserID != userID {
		t.Fatalf("ValidateAccessToken() userID = %s, want %s", accessUserID, userID)
	}

	refreshUserID, err := svc.ValidateRefreshToken(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}
	if refreshUserID != userID {
		t.Fatalf("ValidateRefreshToken() userID = %s, want %s", refreshUserID, userID)
	}
}

func TestAuthService_ValidateAccessToken_Revoked(t *testing.T) {
	setAuthConfigForTests()
	storage := &tokenStorageMock{}
	svc := NewAuthService(storage)

	userID := uuid.New()
	accessToken, _, err := svc.GenerateTokens(userID)
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}

	storage.checkRevoked = true
	_, err = svc.ValidateAccessToken(context.Background(), accessToken)
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want revoked error")
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("ValidateAccessToken() error = %v, want contains 'revoked'", err)
	}
}

func TestAuthService_RevokeAuthTokens_BothTokens(t *testing.T) {
	setAuthConfigForTests()
	storage := &tokenStorageMock{}
	svc := NewAuthService(storage)

	userID := uuid.New()
	accessToken, refreshToken, err := svc.GenerateTokens(userID)
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}

	if err = svc.RevokeAuthTokens(context.Background(), accessToken, refreshToken); err != nil {
		t.Fatalf("RevokeAuthTokens() error = %v", err)
	}

	if storage.revokeCalls != 2 {
		t.Fatalf("RevokeAuthTokens() calls = %d, want 2", storage.revokeCalls)
	}
	if len(storage.revokeTTLs) != 2 {
		t.Fatalf("recorded TTLs = %d, want 2", len(storage.revokeTTLs))
	}
	for i, ttl := range storage.revokeTTLs {
		if ttl <= 0 {
			t.Fatalf("revoke TTL[%d] = %v, want > 0", i, ttl)
		}
	}
}
