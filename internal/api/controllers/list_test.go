package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/models"
	svcErr "wishlist/internal/services/errors"
)

type listControllerAuthMock struct {
	validateAccessTokenFn func(ctx context.Context, token string) (uuid.UUID, error)
}

func (m *listControllerAuthMock) ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.validateAccessTokenFn != nil {
		return m.validateAccessTokenFn(ctx, token)
	}
	return uuid.Nil, errors.New("not implemented")
}

type listControllerServiceMock struct {
	createListFn                func(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error)
	getListByIDFn               func(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, error)
	getListBySharedLinkFn       func(ctx context.Context, token string) (models.List, error)
	getListWithWishesFn         func(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, []models.Wish, error)
	getListWithWishesBySharedFn func(ctx context.Context, token string) (models.List, []models.Wish, error)
	getCurrentUserListsFn       func(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	getPublicListsByUserIDFn    func(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	updateListFn                func(ctx context.Context, listID, userID uuid.UUID, req models.UpdateListRequest) error
	rotateSharedLinkFn          func(ctx context.Context, listID, userID uuid.UUID) (string, error)
	deleteListFn                func(ctx context.Context, listID, userID uuid.UUID) error
}

func (m *listControllerServiceMock) CreateList(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error) {
	if m.createListFn != nil {
		return m.createListFn(ctx, userID, req)
	}
	return models.List{}, nil
}

func (m *listControllerServiceMock) GetListByID(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, error) {
	if m.getListByIDFn != nil {
		return m.getListByIDFn(ctx, listID, requestedByUserID)
	}
	return models.List{}, nil
}

func (m *listControllerServiceMock) GetListBySharedLink(ctx context.Context, token string) (models.List, error) {
	if m.getListBySharedLinkFn != nil {
		return m.getListBySharedLinkFn(ctx, token)
	}
	return models.List{}, nil
}

func (m *listControllerServiceMock) GetListWithWishes(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, []models.Wish, error) {
	if m.getListWithWishesFn != nil {
		return m.getListWithWishesFn(ctx, listID, requestedByUserID)
	}
	return models.List{}, nil, nil
}

func (m *listControllerServiceMock) GetListWithWishesBySharedLink(ctx context.Context, token string) (models.List, []models.Wish, error) {
	if m.getListWithWishesBySharedFn != nil {
		return m.getListWithWishesBySharedFn(ctx, token)
	}
	return models.List{}, nil, nil
}

func (m *listControllerServiceMock) GetCurrentUserLists(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	if m.getCurrentUserListsFn != nil {
		return m.getCurrentUserListsFn(ctx, userID)
	}
	return nil, nil
}

func (m *listControllerServiceMock) GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	if m.getPublicListsByUserIDFn != nil {
		return m.getPublicListsByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *listControllerServiceMock) UpdateList(ctx context.Context, listID, userID uuid.UUID, req models.UpdateListRequest) error {
	if m.updateListFn != nil {
		return m.updateListFn(ctx, listID, userID, req)
	}
	return nil
}

func (m *listControllerServiceMock) RotateSharedLink(ctx context.Context, listID, userID uuid.UUID) (string, error) {
	if m.rotateSharedLinkFn != nil {
		return m.rotateSharedLinkFn(ctx, listID, userID)
	}
	return "", nil
}

func (m *listControllerServiceMock) DeleteList(ctx context.Context, listID, userID uuid.UUID) error {
	if m.deleteListFn != nil {
		return m.deleteListFn(ctx, listID, userID)
	}
	return nil
}

func setupListControllerForTest(as *listControllerAuthMock, ls *listControllerServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	viper.Set(config.ApiBasePath, "/api/v1")
	router := gin.New()
	mw := middlewares.NewMiddlewares(as)
	ctrl := NewListsController(router, mw, ls)
	ctrl.RegisterRoutes()
	return router
}

func listJSONRequest(router *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
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

func TestListsController_CreateList(t *testing.T) {
	currentUserID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("bad request", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists", `{"title":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{createListFn: func(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error) {
			return models.List{}, errors.New("db down")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists", `{"title":"A"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now().UTC()
		expected := models.List{
			ID:         uuid.New(),
			UserID:     currentUserID,
			Title:      "Birthday",
			IsPublic:   true,
			ShareToken: "12345678901234567890123456789012",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		ls := &listControllerServiceMock{createListFn: func(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error) {
			if userID != currentUserID || req.Title != "Birthday" {
				t.Fatalf("unexpected create args")
			}
			return expected, nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists", `{"title":"Birthday"}`, "ok")
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
		}

		var response models.ListResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		if response.ID != expected.ID || response.ShareToken != expected.ShareToken {
			t.Fatalf("unexpected response: %+v", response)
		}
	})
}

func TestListsController_GetCurrentUserLists(t *testing.T) {
	currentUserID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{getCurrentUserListsFn: func(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
			return nil, errors.New("db")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ls := &listControllerServiceMock{getCurrentUserListsFn: func(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
			return []models.List{{ID: uuid.New(), UserID: userID, Title: "Mine", IsPublic: true}}, nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists", "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestListsController_GetListByID(t *testing.T) {
	currentUserID := uuid.New()
	listID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/not-uuid", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ls := &listControllerServiceMock{getListWithWishesFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) (models.List, []models.Wish, error) {
			return models.List{}, nil, svcErr.ForbiddenError{Message: "private"}
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/"+listID.String(), "", "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{getListWithWishesFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) (models.List, []models.Wish, error) {
			return models.List{}, nil, errors.New("db")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/"+listID.String(), "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success owner", func(t *testing.T) {
		ls := &listControllerServiceMock{getListWithWishesFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) (models.List, []models.Wish, error) {
			return models.List{ID: listID, UserID: currentUserID, Title: "Mine", IsPublic: true, ShareToken: "token"}, []models.Wish{{ID: uuid.New(), ListID: listID, Title: "Gift"}}, nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/"+listID.String(), "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestListsController_UpdateList(t *testing.T) {
	currentUserID := uuid.New()
	listID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodPatch, "/api/v1/lists/not-uuid", `{"title":"A"}`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("bad request body", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String(), `{"title":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ls := &listControllerServiceMock{updateListFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID, req models.UpdateListRequest) error {
			return svcErr.ForbiddenError{Message: "not owner"}
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String(), `{"title":"A"}`, "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{updateListFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID, req models.UpdateListRequest) error {
			return errors.New("db")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String(), `{"title":"A"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ls := &listControllerServiceMock{updateListFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID, req models.UpdateListRequest) error {
			if gotListID != listID || gotUserID != currentUserID {
				t.Fatalf("unexpected IDs")
			}
			return nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPatch, "/api/v1/lists/"+listID.String(), `{"title":"A"}`, "ok")
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})
}

func TestListsController_RotateSharedLink(t *testing.T) {
	currentUserID := uuid.New()
	listID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists/not-uuid/rotate-share-link", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ls := &listControllerServiceMock{rotateSharedLinkFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) (string, error) {
			return "", svcErr.ForbiddenError{Message: "not owner"}
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/rotate-share-link", "", "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{rotateSharedLinkFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) (string, error) {
			return "", errors.New("db")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/rotate-share-link", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ls := &listControllerServiceMock{rotateSharedLinkFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) (string, error) {
			return "new-token", nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodPost, "/api/v1/lists/"+listID.String()+"/rotate-share-link", "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestListsController_DeleteList(t *testing.T) {
	currentUserID := uuid.New()
	listID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("invalid list ID", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodDelete, "/api/v1/lists/not-uuid", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ls := &listControllerServiceMock{deleteListFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) error {
			return svcErr.ForbiddenError{Message: "not owner"}
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String(), "", "ok")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{deleteListFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) error {
			return errors.New("db")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String(), "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		called := false
		ls := &listControllerServiceMock{deleteListFn: func(ctx context.Context, gotListID, gotUserID uuid.UUID) error {
			called = true
			return nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodDelete, "/api/v1/lists/"+listID.String(), "", "ok")
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
		if !called {
			t.Fatal("DeleteList service was not called")
		}
	})
}

func TestListsController_GetListBySharedLink(t *testing.T) {
	slug := "12345678901234567890123456789012"
	ownerID := uuid.New()
	listID := uuid.New()

	t.Run("invalid slug", func(t *testing.T) {
		router := setupListControllerForTest(&listControllerAuthMock{}, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/shared/short", "", "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{getListWithWishesBySharedFn: func(ctx context.Context, token string) (models.List, []models.Wish, error) {
			return models.List{}, nil, errors.New("db")
		}}
		router := setupListControllerForTest(&listControllerAuthMock{}, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/shared/"+slug, "", "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success guest", func(t *testing.T) {
		ls := &listControllerServiceMock{getListWithWishesBySharedFn: func(ctx context.Context, token string) (models.List, []models.Wish, error) {
			return models.List{ID: listID, UserID: ownerID, Title: "Shared", IsPublic: true}, []models.Wish{{ID: uuid.New(), ListID: listID, Title: "Gift"}}, nil
		}}
		router := setupListControllerForTest(&listControllerAuthMock{}, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/shared/"+slug, "", "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("success owner", func(t *testing.T) {
		as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return ownerID, nil }}
		ls := &listControllerServiceMock{getListWithWishesBySharedFn: func(ctx context.Context, token string) (models.List, []models.Wish, error) {
			return models.List{ID: listID, UserID: ownerID, Title: "Shared", IsPublic: true, ShareToken: slug}, []models.Wish{{ID: uuid.New(), ListID: listID, Title: "Gift"}}, nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/lists/shared/"+slug, "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestListsController_GetPublicListsByUserID(t *testing.T) {
	currentUserID := uuid.New()
	targetUserID := uuid.New()
	as := &listControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return currentUserID, nil }}

	t.Run("invalid user ID", func(t *testing.T) {
		router := setupListControllerForTest(as, &listControllerServiceMock{})
		w := listJSONRequest(router, http.MethodGet, "/api/v1/users/not-uuid/lists", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ls := &listControllerServiceMock{getPublicListsByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
			return nil, errors.New("db")
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/users/"+targetUserID.String()+"/lists", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		ls := &listControllerServiceMock{getPublicListsByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
			return []models.List{
				{ID: uuid.New(), UserID: targetUserID, Title: "Public", IsPublic: true},
				{ID: uuid.New(), UserID: currentUserID, Title: "Mine", IsPublic: true, ShareToken: "token"},
			}, nil
		}}
		router := setupListControllerForTest(as, ls)
		w := listJSONRequest(router, http.MethodGet, "/api/v1/users/"+targetUserID.String()+"/lists", "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}
