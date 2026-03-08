package controllers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/models"
	svcErr "wishlist/internal/services/errors"
)

type wishControllerAuthMock struct {
	validateAccessTokenFn func(ctx context.Context, token string) (uuid.UUID, error)
}

func (m *wishControllerAuthMock) ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.validateAccessTokenFn != nil {
		return m.validateAccessTokenFn(ctx, token)
	}
	return uuid.Nil, errors.New("not implemented")
}

type wishControllerServiceMock struct {
	createWishFn  func(ctx context.Context, listID, userID uuid.UUID, req models.CreateWishRequest) (models.Wish, error)
	getWishByIDFn func(ctx context.Context, wishID uuid.UUID) (models.Wish, error)
	updateWishFn  func(ctx context.Context, listID, wishID, userID uuid.UUID, req models.UpdateWishRequest) error
	reserveWishFn func(ctx context.Context, listID, wishID, userID uuid.UUID) error
	releaseWishFn func(ctx context.Context, listID, wishID, userID uuid.UUID) error
	deleteWishFn  func(ctx context.Context, listID, wishID, userID uuid.UUID) error
}

func (m *wishControllerServiceMock) CreateWish(ctx context.Context, listID, userID uuid.UUID, req models.CreateWishRequest) (models.Wish, error) {
	if m.createWishFn != nil {
		return m.createWishFn(ctx, listID, userID, req)
	}
	return models.Wish{}, nil
}

func (m *wishControllerServiceMock) GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error) {
	if m.getWishByIDFn != nil {
		return m.getWishByIDFn(ctx, wishID)
	}
	return models.Wish{}, nil
}

func (m *wishControllerServiceMock) UpdateWish(ctx context.Context, listID, wishID, userID uuid.UUID, req models.UpdateWishRequest) error {
	if m.updateWishFn != nil {
		return m.updateWishFn(ctx, listID, wishID, userID, req)
	}
	return nil
}

func (m *wishControllerServiceMock) ReserveWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	if m.reserveWishFn != nil {
		return m.reserveWishFn(ctx, listID, wishID, userID)
	}
	return nil
}

func (m *wishControllerServiceMock) ReleaseWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	if m.releaseWishFn != nil {
		return m.releaseWishFn(ctx, listID, wishID, userID)
	}
	return nil
}

func (m *wishControllerServiceMock) DeleteWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	if m.deleteWishFn != nil {
		return m.deleteWishFn(ctx, listID, wishID, userID)
	}
	return nil
}

func setupWishControllerForTest(as *wishControllerAuthMock, ws *wishControllerServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	viper.Set(config.ApiBasePath, "/api/v1")

	router := gin.New()
	mw := middlewares.NewMiddlewares(as)
	ctrl := NewWishesController(router, mw, ws)
	ctrl.RegisterRoutes()
	return router
}

func wishJSONRequest(router *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestWishesController_CreateWish(t *testing.T) {
	userID := uuid.New()
	listID := uuid.New()
	as := &wishControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/not-uuid/wishes", `{"title":"Gift"}`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("bad request body", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes", `{"title":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ws := &wishControllerServiceMock{createWishFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID, req models.CreateWishRequest) (models.Wish, error) {
			return models.Wish{}, svcErr.ForbiddenError{Message: "not owner"}
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes", `{"title":"Gift"}`, "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ws := &wishControllerServiceMock{createWishFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID, req models.CreateWishRequest) (models.Wish, error) {
			return models.Wish{}, errors.New("db")
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes", `{"title":"Gift"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ws := &wishControllerServiceMock{createWishFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID, req models.CreateWishRequest) (models.Wish, error) {
			if gotListID != listID || gotUserID != userID || req.Title != "Gift" {
				t.Fatalf("unexpected create args")
			}
			return models.Wish{ID: uuid.New(), ListID: gotListID, Title: req.Title}, nil
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes", `{"title":"Gift"}`, "ok")
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
		}
	})
}

func TestWishesController_UpdateWish(t *testing.T) {
	userID := uuid.New()
	listID := uuid.New()
	wishID := uuid.New()
	as := &wishControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPatch, "/api/v1/lists/not-uuid/wishes/"+wishID.String(), `{"title":"new"}`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid wish ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String()+"/wishes/not-uuid", `{"title":"new"}`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("bad request body", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), `{"title":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ws := &wishControllerServiceMock{updateWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID, req models.UpdateWishRequest) error {
			return svcErr.ForbiddenError{Message: "not owner"}
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), `{"title":"new"}`, "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal", func(t *testing.T) {
		ws := &wishControllerServiceMock{updateWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID, req models.UpdateWishRequest) error {
			return errors.New("db")
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), `{"title":"new"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ws := &wishControllerServiceMock{updateWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID, req models.UpdateWishRequest) error {
			if gotListID != listID || gotWishID != wishID || gotUserID != userID {
				t.Fatalf("unexpected update args")
			}
			return nil
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), `{"title":"new"}`, "ok")
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})
}

func TestWishesController_ReserveWish(t *testing.T) {
	userID := uuid.New()
	listID := uuid.New()
	wishID := uuid.New()
	as := &wishControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/not-uuid/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid wish ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes/not-uuid/reserve", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation", func(t *testing.T) {
		ws := &wishControllerServiceMock{reserveWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			return svcErr.ValidationError{Message: "already reserved"}
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal", func(t *testing.T) {
		ws := &wishControllerServiceMock{reserveWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			return errors.New("db")
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ws := &wishControllerServiceMock{reserveWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			if gotListID != listID || gotWishID != wishID || gotUserID != userID {
				t.Fatalf("unexpected reserve args")
			}
			return nil
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})
}

func TestWishesController_ReleaseWish(t *testing.T) {
	userID := uuid.New()
	listID := uuid.New()
	wishID := uuid.New()
	as := &wishControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/not-uuid/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid wish ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/not-uuid/reserve", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal", func(t *testing.T) {
		ws := &wishControllerServiceMock{releaseWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			return errors.New("db")
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		called := false
		ws := &wishControllerServiceMock{releaseWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			called = true
			if gotListID != listID || gotWishID != wishID || gotUserID != userID {
				t.Fatalf("unexpected release args")
			}
			return nil
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String()+"/reserve", "", "ok")
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
		if !called {
			t.Fatal("ReleaseWish service was not called")
		}
	})
}

func TestWishesController_DeleteWish(t *testing.T) {
	userID := uuid.New()
	listID := uuid.New()
	wishID := uuid.New()
	as := &wishControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/not-uuid/wishes/"+wishID.String(), "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid wish ID", func(t *testing.T) {
		router := setupWishControllerForTest(as, &wishControllerServiceMock{})
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/not-uuid", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ws := &wishControllerServiceMock{deleteWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			return svcErr.ForbiddenError{Message: "not owner"}
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), "", "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal", func(t *testing.T) {
		ws := &wishControllerServiceMock{deleteWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			return errors.New("db")
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		called := false
		ws := &wishControllerServiceMock{deleteWishFn: func(ctx context.Context, gotListID, gotWishID, gotUserID uuid.UUID) error {
			called = true
			if gotListID != listID || gotWishID != wishID || gotUserID != userID {
				t.Fatalf("unexpected delete args")
			}
			return nil
		}}
		router := setupWishControllerForTest(as, ws)
		w := wishJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String()+"/wishes/"+wishID.String(), "", "ok")
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
		if !called {
			t.Fatal("DeleteWish service was not called")
		}
	})
}
