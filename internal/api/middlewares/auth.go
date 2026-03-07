package middlewares

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wishlist/internal/api/errors"
)

type AuthService interface {
	ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error)
}

type Middlewares struct {
	authService AuthService
}

func NewMiddlewares(as AuthService) *Middlewares {
	return &Middlewares{authService: as}
}

func (mw *Middlewares) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			apiModels.Error(ctx, 401, "missing authorization header")
			return
		}

		tokenString, found := strings.CutPrefix(header, "Bearer ")
		if !found {
			apiModels.Error(ctx, 401, "invalid authorization format")
			return
		}

		userID, err := mw.authService.ValidateAccessToken(ctx, tokenString)
		if err != nil {
			apiModels.Error(ctx, 401, "invalid or expired token")
			return
		}

		ctx.Set("user_id", userID)
		ctx.Next()
	}
}

func (mw *Middlewares) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.Next()
			return
		}

		tokenString, found := strings.CutPrefix(header, "Bearer ")
		if !found {
			ctx.Next()
			return
		}

		userID, err := mw.authService.ValidateAccessToken(ctx, tokenString)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set("user_id", userID)
		ctx.Next()
	}
}

func GetUserID(ctx *gin.Context) (uuid.UUID, bool) {
	value, exists := ctx.Get("user_id")
	if !exists {
		return uuid.UUID{}, false
	}

	userID, ok := value.(uuid.UUID)
	return userID, ok
}
