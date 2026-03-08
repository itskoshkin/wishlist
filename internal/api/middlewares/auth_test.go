//go:build integration

package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type middlewareAuthServiceMock struct {
	userID uuid.UUID
	err    error
}

func (m *middlewareAuthServiceMock) ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return m.userID, nil
}

func TestAuthMiddleware_SetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expectedUserID := uuid.New()
	mw := NewMiddlewares(&middlewareAuthServiceMock{userID: expectedUserID})
	router := gin.New()

	router.GET("/protected", mw.AuthMiddleware(), func(ctx *gin.Context) {
		userID, ok := GetUserID(ctx)
		if !ok || userID != expectedUserID {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewMiddlewares(&middlewareAuthServiceMock{userID: uuid.New()})
	router := gin.New()
	router.GET("/protected", mw.AuthMiddleware(), func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestOptionalAuthMiddleware_InvalidToken_AllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewMiddlewares(&middlewareAuthServiceMock{err: errors.New("invalid token")})
	router := gin.New()
	router.GET("/public", mw.OptionalAuthMiddleware(), func(ctx *gin.Context) {
		if _, ok := GetUserID(ctx); ok {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestOptionalAuthMiddleware_ValidToken_SetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expectedUserID := uuid.New()
	mw := NewMiddlewares(&middlewareAuthServiceMock{userID: expectedUserID})
	router := gin.New()
	router.GET("/public", mw.OptionalAuthMiddleware(), func(ctx *gin.Context) {
		userID, ok := GetUserID(ctx)
		if !ok || userID != expectedUserID {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
