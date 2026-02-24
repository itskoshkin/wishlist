package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

type AuthServiceImpl struct {
	accessTokenSecret  string
	refreshTokenSecret string
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
}

func NewAuthService() *AuthServiceImpl {
	return &AuthServiceImpl{
		accessTokenSecret:  viper.GetString(config.AccessTokenSecret),
		refreshTokenSecret: viper.GetString(config.RefreshTokenSecret),
		accessTokenTTL:     viper.GetDuration(config.AccessTokenTTL),
		refreshTokenTTL:    viper.GetDuration(config.RefreshTokenTTL),
	}
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func (svc *AuthServiceImpl) GenerateTokens(userID uuid.UUID) (access, refresh string, err error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID.String(),
			Issuer:    "itskoshkin/wishlist",
			Audience:  jwt.ClaimStrings{"Wishlist API"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(svc.accessTokenTTL)),
		},
	})
	signedAccessToken, err := accessToken.SignedString([]byte(svc.accessTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID.String(),
			Issuer:    "itskoshkin/wishlist",
			Audience:  jwt.ClaimStrings{"Wishlist API"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(svc.refreshTokenTTL)),
		},
	})
	signedRefreshToken, err := refreshToken.SignedString([]byte(svc.refreshTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (svc *AuthServiceImpl) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	return svc.validateToken(tokenString, svc.accessTokenSecret)
}

func (svc *AuthServiceImpl) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	return svc.validateToken(tokenString, svc.refreshTokenSecret)
}

func (svc *AuthServiceImpl) validateToken(tokenString, secretString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretString), nil
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}
