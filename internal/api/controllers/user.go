package controllers

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/api/errors"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/models"
)

type AuthService interface {
	GenerateTokens(userID uuid.UUID) (string, string, error)
	ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error)
	ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	RevokeAuthTokens(ctx context.Context, accessToken, refreshToken string) error
}

type UserService interface {
	Register(ctx context.Context, req models.RegisterUserRequest) (models.User, error)
	VerifyEmail(ctx context.Context, token string) error
	LogIn(ctx context.Context, req models.LogInUserRequest) (models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	SearchUsersByUsername(ctx context.Context, query string, limit int) ([]models.User, error)
	UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error
	DeleteAvatar(ctx context.Context, id uuid.UUID) error
	VerifyPassword(ctx context.Context, id uuid.UUID, password string) error
	ChangePassword(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UsersController struct {
	router      *gin.Engine
	mw          *middlewares.Middlewares
	authService AuthService
	userService UserService
}

func NewUsersController(e *gin.Engine, mw *middlewares.Middlewares, as AuthService, us UserService) *UsersController {
	return &UsersController{router: e, mw: mw, authService: as, userService: us}
}

func (ctrl *UsersController) RegisterRoutes() {
	basePath := ctrl.router.Group(viper.GetString(config.ApiBasePath))
	authRoutes := basePath.Group("/auth")
	{
		authRoutes.POST("/register", ctrl.Register)
		authRoutes.POST("/verify-email", ctrl.VerifyEmail)
		authRoutes.POST("/login", ctrl.LogIn)
		authRoutes.POST("/refresh", ctrl.RefreshTokens)
		authRoutes.POST("/logout", ctrl.mw.AuthMiddleware(), ctrl.LogOut)

		authRoutes.POST("/forgot-password", ctrl.ForgotPassword)
		authRoutes.POST("/set-new-password", ctrl.SetNewPassword)
	}
	userRoutes := basePath.Group("/users")
	{
		authedUserRoutes := userRoutes.Group("").Use(ctrl.mw.AuthMiddleware())
		{
			authedUserRoutes.GET("/me", ctrl.GetCurrentUser)
			authedUserRoutes.PATCH("/me", ctrl.UpdateCurrentUser)
			authedUserRoutes.PUT("/me/avatar", ctrl.UpdateAvatar)
			authedUserRoutes.DELETE("/me/avatar", ctrl.DeleteAvatar)
			authedUserRoutes.PATCH("/me/update-password", ctrl.UpdateCurrentPassword)
			authedUserRoutes.DELETE("/me", ctrl.DeleteCurrentUser)

			authedUserRoutes.GET("/search", ctrl.SearchUsers)
			authedUserRoutes.GET("/by-username/:username", ctrl.GetUserByUsername)
			authedUserRoutes.GET("/:user_id", ctrl.GetUserByID)
		}
	}
}

// Register GoDoc
// @Summary Init user registration
// @Description Register a new account and return auth tokens; if email is provided, verification flow is also initiated
// @Tags auth
// @Accept json
// @Produce json
// @Param RegisterRequest body models.RegisterUserRequest true "Registration payload"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/register [post]
func (ctrl *UsersController) Register(ctx *gin.Context) {
	var req models.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	user, err := ctrl.userService.Register(ctx, req)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	accessToken, refreshToken, err := ctrl.authService.GenerateTokens(user.ID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, models.AuthResponse{
		AuthTokensResponse: models.AuthTokensResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		User: user.ToPrivateResponse(),
	})
}

// VerifyEmail GoDoc
// @Summary Verify email
// @Description Verify user email by token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.VerifyEmailRequest true "Verification token"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/verify-email [post]
func (ctrl *UsersController) VerifyEmail(ctx *gin.Context) {
	var req models.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.VerifyEmail(ctx, req.Token); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "email verified"})
}

// LogIn GoDoc
// @Summary Login user
// @Description Login with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LogInUserRequest true "Credentials"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/login [post]
func (ctrl *UsersController) LogIn(ctx *gin.Context) {
	var req models.LogInUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	user, err := ctrl.userService.LogIn(ctx, req)
	if err != nil {
		apiModels.Error(ctx, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, refreshToken, err := ctrl.authService.GenerateTokens(user.ID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, models.AuthResponse{
		AuthTokensResponse: models.AuthTokensResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		User: user.ToPrivateResponse(),
	})
}

// RefreshTokens GoDoc
// @Summary Refresh tokens
// @Description Get new access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.AuthTokensResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/refresh [post]
func (ctrl *UsersController) RefreshTokens(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	userID, err := ctrl.authService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		apiModels.Error(ctx, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	accessToken, refreshToken, err := ctrl.authService.GenerateTokens(userID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, models.AuthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// LogOut GoDoc
// @Summary Logout user
// @Description Revoke access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/logout [post]
func (ctrl *UsersController) LogOut(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	accessToken, found := strings.CutPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if !found {
		apiModels.Error(ctx, http.StatusUnauthorized, "invalid or expired access token")
		return
	}

	if err := ctrl.authService.RevokeAuthTokens(ctx, accessToken, req.RefreshToken); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "logged out"})
}

// GetCurrentUser GoDoc
// @Summary Get current user
// @Description Get current authenticated user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me [get]
func (ctrl *UsersController) GetCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, user.ToPrivateResponse())
}

// GetUserByID GoDoc
// @Summary Get user by ID
// @Description Get user public profile by ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/{user_id} [get]
func (ctrl *UsersController) GetUserByID(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, user.ToPublicResponse())
}

// GetUserByUsername GoDoc
// @Summary Get user by username
// @Description Get user public profile by username
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 404 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/by-username/{username} [get]
func (ctrl *UsersController) GetUserByUsername(ctx *gin.Context) {
	username := strings.TrimSpace(ctx.Param("username"))
	if username == "" {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid username")
		return
	}

	user, err := ctrl.userService.GetUserByUsername(ctx, username)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, user.ToPublicResponse())
}

func (ctrl *UsersController) SearchUsers(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("query"))
	if len(query) < 2 {
		apiModels.Error(ctx, http.StatusBadRequest, "search query too short")
		return
	}

	users, err := ctrl.userService.SearchUsersByUsername(ctx, query, 8)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	resp := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, user.ToPublicResponse())
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateCurrentUser GoDoc
// @Summary Update current user
// @Description Update current authenticated user fields
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateUserRequest true "Update payload"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me [patch]
// noinspection DuplicatedCode
func (ctrl *UsersController) UpdateCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.UpdateUserByID(ctx, userID, req); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, user.ToPrivateResponse())
}

// UpdateAvatar GoDoc
// @Summary Update user avatar
// @Description Upload new avatar for current user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me/avatar [put]
func (ctrl *UsersController) UpdateAvatar(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	//file, header, err := ctx.Request.FormFile("avatar")
	fileHeader, err := ctx.FormFile("avatar")
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "avatar file is required")
		return
	}

	maxSize := int64(viper.GetInt(config.MinioMaxFileSize)) << 20 // 5 << 20 = 5 * 1 048 576 = 5 242 880 bytes = 5 MB
	if fileHeader.Size > maxSize {
		apiModels.Error(ctx, http.StatusBadRequest, "avatar file too large (max 10 MB)")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	switch contentType {
	case /*"image/heic", */ "image/jpeg", "image/png", "image/webp", "image/gif": //TODO: HEIC isn't native for browsers except Safari, will need to use some lib to convert to JPG
		// continue
	default:
		apiModels.Error(ctx, http.StatusBadRequest, "unsupported image format (PNG, JPG, WEBP or GIF only)") // HEIC
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	if err = ctrl.userService.UpdateAvatar(ctx, userID, file, fileHeader.Size, contentType); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, user.ToPrivateResponse())
}

// DeleteAvatar GoDoc
// @Summary Delete user avatar
// @Description Delete current user avatar
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} apiModels.APIResponse
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me/avatar [delete]
func (ctrl *UsersController) DeleteAvatar(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := ctrl.userService.DeleteAvatar(ctx, userID); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.APIResponse{Message: "avatar deleted"})
}

// UpdateCurrentPassword GoDoc
// @Summary Change current password
// @Description Change password for current authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ChangePasswordRequest true "Current and new password"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me/update-password [patch]
// noinspection DuplicatedCode
func (ctrl *UsersController) UpdateCurrentPassword(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.ChangePassword(ctx, userID, req); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "password changed"})
}

// ForgotPassword GoDoc
// @Summary Request password reset
// @Description Send password reset link if account exists
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Email"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/forgot-password [post]
func (ctrl *UsersController) ForgotPassword(ctx *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.RequestPasswordReset(ctx, req.Email); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "if account with this email exists, you will receive a password reset link shortly"})
}

// SetNewPassword GoDoc
// @Summary Set new password
// @Description Set new password by reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.SetNewPasswordRequest true "Token and new password"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/set-new-password [post]
func (ctrl *UsersController) SetNewPassword(ctx *gin.Context) {
	var req models.SetNewPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "password has been reset"})
}

// DeleteCurrentUser GoDoc
// @Summary Delete current user
// @Description Delete current user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.DeleteAccountRequest true "Current password"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me [delete]
func (ctrl *UsersController) DeleteCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.DeleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.VerifyPassword(ctx, userID, req.CurrentPassword); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	if err := ctrl.userService.Delete(ctx, userID); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "account deleted"})
}
