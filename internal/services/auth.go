package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

type TokenStorage interface {
	SaveEmailVerificationToken(ctx context.Context, tokenID, userID string) error
	GetEmailVerificationToken(ctx context.Context, tokenID string) (string, error)
	DeleteEmailVerificationToken(ctx context.Context, tokenID string) error
	CheckIfAuthTokenRevoked(ctx context.Context, tokenID string) (bool, error)
	RevokeAuthTokens(ctx context.Context, tokenID string, remainingTTL time.Duration) error
	SavePasswordResetToken(ctx context.Context, tokenID string, userID string) error
	GetPasswordResetToken(ctx context.Context, tokenID string) (string, error)
	DeletePasswordResetToken(ctx context.Context, tokenID string) error
}

type AuthServiceImpl struct {
	accessTokenSecret  string
	refreshTokenSecret string
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
	tokenStorage       TokenStorage
}

func NewAuthService(ts TokenStorage) *AuthServiceImpl {
	return &AuthServiceImpl{
		accessTokenSecret:  viper.GetString(config.AccessTokenSecret),
		refreshTokenSecret: viper.GetString(config.RefreshTokenSecret),
		accessTokenTTL:     viper.GetDuration(config.AccessTokenTTL),
		refreshTokenTTL:    viper.GetDuration(config.RefreshTokenTTL),
		tokenStorage:       ts,
	}
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func (svc *AuthServiceImpl) GenerateTokens(userID uuid.UUID) (access, refresh string, err error) {
	accessToken := issueNewToken(userID, svc.accessTokenTTL)
	signedAccessToken, err := accessToken.SignedString([]byte(svc.accessTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := issueNewToken(userID, svc.refreshTokenTTL)
	signedRefreshToken, err := refreshToken.SignedString([]byte(svc.refreshTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedAccessToken, signedRefreshToken, nil
}

func issueNewToken(userID uuid.UUID, ttl time.Duration) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID.String(),
			Issuer:    viper.GetString(config.JwtIssuer),
			Audience:  viper.GetStringSlice(config.JwtAudience),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	})
}

// noinspection DuplicatedCode
func (svc *AuthServiceImpl) ValidateAccessToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims, err := svc.validateToken(tokenString, svc.accessTokenSecret)
	if err != nil {
		return uuid.Nil, err
	}

	revoked, err := svc.tokenStorage.CheckIfAuthTokenRevoked(ctx, claims.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check token revocation: %w", err)
	}

	if revoked {
		return uuid.Nil, fmt.Errorf("token has been revoked")
	}

	return claims.UserID, nil
}

// noinspection DuplicatedCode
func (svc *AuthServiceImpl) ValidateRefreshToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims, err := svc.validateToken(tokenString, svc.refreshTokenSecret)
	if err != nil {
		return uuid.Nil, err
	}

	revoked, err := svc.tokenStorage.CheckIfAuthTokenRevoked(ctx, claims.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check token revocation: %w", err)
	}

	if revoked {
		return uuid.Nil, fmt.Errorf("token has been revoked")
	}

	return claims.UserID, nil
}

func (svc *AuthServiceImpl) validateToken(tokenString, secretString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretString), nil
	}, jwt.WithAudience(viper.GetStringSlice(config.JwtAudience)...))
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (svc *AuthServiceImpl) RevokeAuthTokens(ctx context.Context, accessToken, refreshToken string) error {
	errs := make([]error, 0)

	if claims, err := svc.validateToken(accessToken, svc.accessTokenSecret); err == nil {
		if remaining := time.Until(claims.ExpiresAt.Time); remaining > 0 {
			if err = svc.tokenStorage.RevokeAuthTokens(ctx, claims.ID, remaining); err != nil {
				errs = append(errs, fmt.Errorf("access token: %w", err))
			}
		}
	}

	if claims, err := svc.validateToken(refreshToken, svc.refreshTokenSecret); err == nil {
		if remaining := time.Until(claims.ExpiresAt.Time); remaining > 0 {
			if err = svc.tokenStorage.RevokeAuthTokens(ctx, claims.ID, remaining); err != nil {
				errs = append(errs, fmt.Errorf("refresh token: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to revoke tokens: %v", errs)
	}

	return nil
}
