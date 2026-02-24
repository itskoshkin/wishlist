package storage

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

const (
	projectPrefix       = "wishlist"
	revokedTokensPrefix = projectPrefix + ":" + "revoked_token:"
	passwordResetPrefix = projectPrefix + ":" + "password_reset_token:"
)

type TokenStorageImpl struct {
	client *redis.Client
	pwdTTL time.Duration // Password Reset Token TTL
}

func NewTokenStorage(client *redis.Client) *TokenStorageImpl {
	return &TokenStorageImpl{client: client, pwdTTL: viper.GetDuration(config.PwdResetTokenTTL)}
}

func (ts *TokenStorageImpl) CheckIfAuthTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	if err := ts.client.Get(ctx, revokedTokensPrefix+tokenID).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ts *TokenStorageImpl) RevokeAuthTokens(ctx context.Context, tokenID string, remainingTTL time.Duration) error {
	return ts.client.Set(ctx, revokedTokensPrefix+tokenID, "ACTIVE, REVOKED", remainingTTL).Err()
}

func (ts *TokenStorageImpl) SavePasswordResetToken(ctx context.Context, tokenID string, userID string) error {
	return ts.client.Set(ctx, passwordResetPrefix+tokenID, userID, ts.pwdTTL).Err()
}

func (ts *TokenStorageImpl) GetPasswordResetToken(ctx context.Context, tokenID string) (string, error) {
	return ts.client.Get(ctx, passwordResetPrefix+tokenID).Result()
}

func (ts *TokenStorageImpl) DeletePasswordResetToken(ctx context.Context, tokenID string) error {
	return ts.client.Del(ctx, passwordResetPrefix+tokenID).Err()
}
