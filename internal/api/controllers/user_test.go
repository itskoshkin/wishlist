package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	apiModels "wishlist/internal/api/errors"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/models"
	svcErr "wishlist/internal/services/errors"
)

type userControllerAuthMock struct {
	generateTokensFn      func(userID uuid.UUID) (string, string, error)
	validateAccessTokenFn func(ctx context.Context, token string) (uuid.UUID, error)
	validateRefreshFn     func(ctx context.Context, token string) (uuid.UUID, error)
	revokeAuthTokensFn    func(ctx context.Context, accessToken, refreshToken string) error
}

func (m *userControllerAuthMock) GenerateTokens(userID uuid.UUID) (string, string, error) {
	if m.generateTokensFn != nil {
		return m.generateTokensFn(userID)
	}
	return "", "", nil
}

func (m *userControllerAuthMock) ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.validateAccessTokenFn != nil {
		return m.validateAccessTokenFn(ctx, token)
	}
	return uuid.Nil, errors.New("not implemented")
}

func (m *userControllerAuthMock) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.validateRefreshFn != nil {
		return m.validateRefreshFn(ctx, token)
	}
	return uuid.Nil, errors.New("not implemented")
}

func (m *userControllerAuthMock) RevokeAuthTokens(ctx context.Context, accessToken, refreshToken string) error {
	if m.revokeAuthTokensFn != nil {
		return m.revokeAuthTokensFn(ctx, accessToken, refreshToken)
	}
	return nil
}

type userControllerServiceMock struct {
	registerFn              func(ctx context.Context, req models.RegisterUserRequest) (models.User, error)
	verifyEmailFn           func(ctx context.Context, token string) error
	logInFn                 func(ctx context.Context, req models.LogInUserRequest) (models.User, error)
	getUserByIDFn           func(ctx context.Context, id uuid.UUID) (models.User, error)
	getUserByUsernameFn     func(ctx context.Context, username string) (models.User, error)
	searchUsersByUsernameFn func(ctx context.Context, query string, limit int) ([]models.User, error)
	updateUserByIDFn        func(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error
	updateAvatarFn          func(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error
	deleteAvatarFn          func(ctx context.Context, id uuid.UUID) error
	verifyPasswordFn        func(ctx context.Context, id uuid.UUID, password string) error
	changePasswordFn        func(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error
	requestPasswordResetFn  func(ctx context.Context, email string) error
	resetPasswordFn         func(ctx context.Context, token, newPassword string) error
	deleteFn                func(ctx context.Context, id uuid.UUID) error
}

func (m *userControllerServiceMock) Register(ctx context.Context, req models.RegisterUserRequest) (models.User, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, req)
	}
	return models.User{}, nil
}

func (m *userControllerServiceMock) VerifyEmail(ctx context.Context, token string) error {
	if m.verifyEmailFn != nil {
		return m.verifyEmailFn(ctx, token)
	}
	return nil
}

func (m *userControllerServiceMock) LogIn(ctx context.Context, req models.LogInUserRequest) (models.User, error) {
	if m.logInFn != nil {
		return m.logInFn(ctx, req)
	}
	return models.User{}, nil
}

func (m *userControllerServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return models.User{}, nil
}

func (m *userControllerServiceMock) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	if m.getUserByUsernameFn != nil {
		return m.getUserByUsernameFn(ctx, username)
	}
	return models.User{}, nil
}

func (m *userControllerServiceMock) SearchUsersByUsername(ctx context.Context, query string, limit int) ([]models.User, error) {
	if m.searchUsersByUsernameFn != nil {
		return m.searchUsersByUsernameFn(ctx, query, limit)
	}
	return nil, nil
}

func (m *userControllerServiceMock) UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
	if m.updateUserByIDFn != nil {
		return m.updateUserByIDFn(ctx, id, req)
	}
	return nil
}

func (m *userControllerServiceMock) UpdateAvatar(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error {
	if m.updateAvatarFn != nil {
		return m.updateAvatarFn(ctx, id, reader, size, contentType)
	}
	return nil
}

func (m *userControllerServiceMock) DeleteAvatar(ctx context.Context, id uuid.UUID) error {
	if m.deleteAvatarFn != nil {
		return m.deleteAvatarFn(ctx, id)
	}
	return nil
}

func (m *userControllerServiceMock) VerifyPassword(ctx context.Context, id uuid.UUID, password string) error {
	if m.verifyPasswordFn != nil {
		return m.verifyPasswordFn(ctx, id, password)
	}
	return nil
}

func (m *userControllerServiceMock) ChangePassword(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error {
	if m.changePasswordFn != nil {
		return m.changePasswordFn(ctx, id, req)
	}
	return nil
}

func (m *userControllerServiceMock) RequestPasswordReset(ctx context.Context, email string) error {
	if m.requestPasswordResetFn != nil {
		return m.requestPasswordResetFn(ctx, email)
	}
	return nil
}

func (m *userControllerServiceMock) ResetPassword(ctx context.Context, token, newPassword string) error {
	if m.resetPasswordFn != nil {
		return m.resetPasswordFn(ctx, token, newPassword)
	}
	return nil
}

func (m *userControllerServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func setupUserControllerForTest(as *userControllerAuthMock, us *userControllerServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	viper.Set(config.ApiBasePath, "/api/v1")
	viper.Set(config.MinioMaxFileSize, 10)

	router := gin.New()
	mw := middlewares.NewMiddlewares(as)
	ctrl := NewUsersController(router, mw, as, us)
	ctrl.RegisterRoutes()
	return router
}

func userJSONRequest(router *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
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

func newAvatarMultipartBody(t *testing.T, filename, contentType string, payload []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	headers := textproto.MIMEHeader{}
	headers.Set("Content-Disposition", `form-data; name="avatar"; filename="`+filename+`"`)
	headers.Set("Content-Type", contentType)

	part, err := writer.CreatePart(headers)
	if err != nil {
		t.Fatalf("CreatePart() error = %v", err)
	}
	if _, err = part.Write(payload); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	return body, writer.FormDataContentType()
}

func TestUsersController_Register(t *testing.T) {
	user := models.User{ID: uuid.New(), Name: "Alice", Username: "alice", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(&userControllerAuthMock{}, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/register", `{"name":`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal register error", func(t *testing.T) {
		us := &userControllerServiceMock{registerFn: func(ctx context.Context, req models.RegisterUserRequest) (models.User, error) {
			return models.User{}, errors.New("db down")
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/register", `{"name":"Alice","username":"alice","password":"password123"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("internal token error", func(t *testing.T) {
		as := &userControllerAuthMock{generateTokensFn: func(userID uuid.UUID) (string, string, error) {
			return "", "", errors.New("jwt down")
		}}
		us := &userControllerServiceMock{registerFn: func(ctx context.Context, req models.RegisterUserRequest) (models.User, error) {
			return user, nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/register", `{"name":"Alice","username":"alice","password":"password123"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		as := &userControllerAuthMock{generateTokensFn: func(userID uuid.UUID) (string, string, error) {
			if userID != user.ID {
				t.Fatalf("GenerateTokens user ID = %s, want %s", userID, user.ID)
			}
			return "access-token", "refresh-token", nil
		}}
		us := &userControllerServiceMock{registerFn: func(ctx context.Context, req models.RegisterUserRequest) (models.User, error) {
			return user, nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/register", `{"name":"Alice","username":"alice","password":"password123"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var resp models.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		if resp.AccessToken != "access-token" || resp.RefreshToken != "refresh-token" || resp.User.ID != user.ID {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})
}

func TestUsersController_VerifyEmail(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(&userControllerAuthMock{}, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/verify-email", `{"token":`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		us := &userControllerServiceMock{verifyEmailFn: func(ctx context.Context, token string) error {
			return svcErr.ValidationError{Message: "invalid token"}
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/verify-email", `{"token":"bad"}`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
		var apiErr apiModels.APIError
		if err := json.Unmarshal(w.Body.Bytes(), &apiErr); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		if apiErr.Message != "invalid token" {
			t.Fatalf("message = %q, want invalid token", apiErr.Message)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		us := &userControllerServiceMock{verifyEmailFn: func(ctx context.Context, token string) error {
			return errors.New("db")
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/verify-email", `{"token":"ok"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{verifyEmailFn: func(ctx context.Context, token string) error { return nil }}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/verify-email", `{"token":"ok"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_LogIn(t *testing.T) {
	user := models.User{ID: uuid.New(), Username: "john", Name: "John", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(&userControllerAuthMock{}, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/login", `{"username":`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		us := &userControllerServiceMock{logInFn: func(ctx context.Context, req models.LogInUserRequest) (models.User, error) {
			return models.User{}, errors.New("bad creds")
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/login", `{"username":"john","password":"bad"}`, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("internal token error", func(t *testing.T) {
		as := &userControllerAuthMock{generateTokensFn: func(userID uuid.UUID) (string, string, error) {
			return "", "", errors.New("jwt")
		}}
		us := &userControllerServiceMock{logInFn: func(ctx context.Context, req models.LogInUserRequest) (models.User, error) {
			return user, nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/login", `{"username":"john","password":"password123"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		as := &userControllerAuthMock{generateTokensFn: func(userID uuid.UUID) (string, string, error) {
			return "a1", "r1", nil
		}}
		us := &userControllerServiceMock{logInFn: func(ctx context.Context, req models.LogInUserRequest) (models.User, error) {
			return user, nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/login", `{"username":"john","password":"password123"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_RefreshTokens(t *testing.T) {
	userID := uuid.New()

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(&userControllerAuthMock{}, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		as := &userControllerAuthMock{validateRefreshFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("invalid")
		}}
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"bad"}`, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("internal token error", func(t *testing.T) {
		as := &userControllerAuthMock{
			validateRefreshFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil },
			generateTokensFn:  func(userID uuid.UUID) (string, string, error) { return "", "", errors.New("jwt") },
		}
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"ok"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		as := &userControllerAuthMock{
			validateRefreshFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil },
			generateTokensFn:  func(userID uuid.UUID) (string, string, error) { return "a2", "r2", nil },
		}
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"ok"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_LogOut(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("middleware unauthorized", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(`{"refresh_token":"r1"}`))
		req.Header.Set("Authorization", "Token wrong")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		asLocal := &userControllerAuthMock{
			validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil },
			revokeAuthTokensFn:    func(ctx context.Context, accessToken, refreshToken string) error { return errors.New("redis") },
		}
		router := setupUserControllerForTest(asLocal, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"r1"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		asLocal := &userControllerAuthMock{
			validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil },
			revokeAuthTokensFn: func(ctx context.Context, accessToken, refreshToken string) error {
				if accessToken != "ok" || refreshToken != "r1" {
					t.Fatalf("unexpected token pair")
				}
				return nil
			},
		}
		router := setupUserControllerForTest(asLocal, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"r1"}`, "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_ForgotPassword(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(&userControllerAuthMock{}, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/forgot-password", `{"email":`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal", func(t *testing.T) {
		us := &userControllerServiceMock{requestPasswordResetFn: func(ctx context.Context, email string) error {
			return errors.New("smtp")
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/forgot-password", `{"email":"alice@example.com"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		calledWith := ""
		us := &userControllerServiceMock{requestPasswordResetFn: func(ctx context.Context, email string) error {
			calledWith = email
			return nil
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/forgot-password", `{"email":"alice@example.com"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if calledWith != "alice@example.com" {
			t.Fatalf("called with = %s", calledWith)
		}
	})
}

func TestUsersController_SetNewPassword(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(&userControllerAuthMock{}, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/set-new-password", `{"token":`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal", func(t *testing.T) {
		us := &userControllerServiceMock{resetPasswordFn: func(ctx context.Context, token, newPassword string) error {
			return errors.New("invalid")
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/set-new-password", `{"token":"t1","new_password":"new12345"}`, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{resetPasswordFn: func(ctx context.Context, token, newPassword string) error {
			return nil
		}}
		router := setupUserControllerForTest(&userControllerAuthMock{}, us)
		w := userJSONRequest(router, http.MethodPost, "/api/v1/auth/set-new-password", `{"token":"t1","new_password":"new12345"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_GetCurrentUser(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("not found", func(t *testing.T) {
		us := &userControllerServiceMock{getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
			return models.User{}, svcErr.NotFoundError{Entity: "user", Field: "id", Value: id.String()}
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/me", "", "ok")
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("internal", func(t *testing.T) {
		us := &userControllerServiceMock{getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
			return models.User{}, errors.New("db")
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/me", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
			return models.User{ID: id, Username: "john", Name: "John"}, nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/me", "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_UpdateCurrentUser(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me", `{"username":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal update", func(t *testing.T) {
		us := &userControllerServiceMock{updateUserByIDFn: func(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
			return errors.New("db")
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me", `{"username":"updated"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("internal get", func(t *testing.T) {
		us := &userControllerServiceMock{
			updateUserByIDFn: func(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error { return nil },
			getUserByIDFn:    func(ctx context.Context, id uuid.UUID) (models.User, error) { return models.User{}, errors.New("db") },
		}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me", `{"username":"updated"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{
			updateUserByIDFn: func(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
				if id != userID {
					t.Fatalf("id mismatch")
				}
				return nil
			},
			getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
				return models.User{ID: id, Username: "updated", Name: "Updated"}, nil
			},
		}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me", `{"username":"updated"}`, "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_UpdateAvatar(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("missing file", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/avatar", nil)
		req.Header.Set("Authorization", "Bearer ok")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		body, contentType := newAvatarMultipartBody(t, "avatar.txt", "text/plain", []byte("hello"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/avatar", body)
		req.Header.Set("Authorization", "Bearer ok")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal update avatar", func(t *testing.T) {
		us := &userControllerServiceMock{updateAvatarFn: func(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error {
			return errors.New("s3")
		}}
		router := setupUserControllerForTest(as, us)
		body, contentType := newAvatarMultipartBody(t, "avatar.png", "image/png", []byte{0x89, 0x50, 0x4e, 0x47})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/avatar", body)
		req.Header.Set("Authorization", "Bearer ok")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("internal get user", func(t *testing.T) {
		us := &userControllerServiceMock{
			updateAvatarFn: func(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error {
				return nil
			},
			getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) { return models.User{}, errors.New("db") },
		}
		router := setupUserControllerForTest(as, us)
		body, contentType := newAvatarMultipartBody(t, "avatar.png", "image/png", []byte{0x89, 0x50, 0x4e, 0x47})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/avatar", body)
		req.Header.Set("Authorization", "Bearer ok")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{
			updateAvatarFn: func(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error {
				if id != userID || contentType != "image/png" {
					t.Fatalf("unexpected update avatar args")
				}
				return nil
			},
			getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
				return models.User{ID: id, Username: "john", Name: "John"}, nil
			},
		}
		router := setupUserControllerForTest(as, us)
		body, contentType := newAvatarMultipartBody(t, "avatar.png", "image/png", []byte{0x89, 0x50, 0x4e, 0x47})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/avatar", body)
		req.Header.Set("Authorization", "Bearer ok")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_DeleteAvatar(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("internal", func(t *testing.T) {
		us := &userControllerServiceMock{deleteAvatarFn: func(ctx context.Context, id uuid.UUID) error { return errors.New("s3") }}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodDelete, "/api/v1/users/me/avatar", "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{deleteAvatarFn: func(ctx context.Context, id uuid.UUID) error {
			if id != userID {
				t.Fatalf("id mismatch")
			}
			return nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodDelete, "/api/v1/users/me/avatar", "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_UpdateCurrentPassword(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me/update-password", `{"current_password":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		us := &userControllerServiceMock{changePasswordFn: func(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error {
			return svcErr.ValidationError{Message: "wrong current password"}
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me/update-password", `{"current_password":"bad","new_password":"new12345"}`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		us := &userControllerServiceMock{changePasswordFn: func(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error {
			return errors.New("db")
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me/update-password", `{"current_password":"old12345","new_password":"new12345"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{changePasswordFn: func(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error {
			if id != userID {
				t.Fatalf("id mismatch")
			}
			return nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodPatch, "/api/v1/users/me/update-password", `{"current_password":"old12345","new_password":"new12345"}`, "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_DeleteCurrentUser(t *testing.T) {
	userID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return userID, nil }}

	t.Run("bad request", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodDelete, "/api/v1/users/me", `{"password":`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		us := &userControllerServiceMock{verifyPasswordFn: func(ctx context.Context, id uuid.UUID, password string) error {
			return svcErr.ValidationError{Message: "wrong password"}
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodDelete, "/api/v1/users/me", `{"password":"bad"}`, "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal delete error", func(t *testing.T) {
		us := &userControllerServiceMock{
			verifyPasswordFn: func(ctx context.Context, id uuid.UUID, password string) error { return nil },
			deleteFn:         func(ctx context.Context, id uuid.UUID) error { return errors.New("db") },
		}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodDelete, "/api/v1/users/me", `{"password":"password123"}`, "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{
			verifyPasswordFn: func(ctx context.Context, id uuid.UUID, password string) error { return nil },
			deleteFn:         func(ctx context.Context, id uuid.UUID) error { return nil },
		}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodDelete, "/api/v1/users/me", `{"password":"password123"}`, "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUsersController_GetUserByID(t *testing.T) {
	targetID := uuid.New()
	as := &userControllerAuthMock{validateAccessTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) { return uuid.New(), nil }}

	t.Run("invalid path param", func(t *testing.T) {
		router := setupUserControllerForTest(as, &userControllerServiceMock{})
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/not-uuid", "", "ok")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		us := &userControllerServiceMock{getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
			return models.User{}, svcErr.NotFoundError{Entity: "user", Field: "id", Value: id.String()}
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/"+targetID.String(), "", "ok")
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		us := &userControllerServiceMock{getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
			return models.User{}, errors.New("db")
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/"+targetID.String(), "", "ok")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success", func(t *testing.T) {
		us := &userControllerServiceMock{getUserByIDFn: func(ctx context.Context, id uuid.UUID) (models.User, error) {
			return models.User{ID: id, Username: "public", Name: "Public"}, nil
		}}
		router := setupUserControllerForTest(as, us)
		w := userJSONRequest(router, http.MethodGet, "/api/v1/users/"+targetID.String(), "", "ok")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}
