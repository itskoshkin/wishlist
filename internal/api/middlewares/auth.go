package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wishlist/internal/api/errors"
)

type AuthService interface {
	ValidateAccessToken(token string) (uuid.UUID, error)
}

type Middlewares struct {
	service AuthService
}

func NewMiddlewares(service AuthService) *Middlewares {
	return &Middlewares{service: service}
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

		userID, err := mw.service.ValidateAccessToken(tokenString)
		if err != nil {
			apiModels.Error(ctx, 401, "invalid or expired token")
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
