package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/api/errors"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/models"
	"wishlist/internal/services/errors"
)

type AuthService interface {
	GenerateTokens(userID uuid.UUID) (string, string, error)
	ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error)
	ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	RevokeAuthTokens(ctx context.Context, accessToken, refreshToken string) error
}

type UserService interface {
	Register(ctx context.Context, req models.RegisterUserRequest) (models.User, error)
	LogIn(ctx context.Context, req models.LogInUserRequest) (models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error
	VerifyPassword(ctx context.Context, id uuid.UUID, password string) error
	ChangePassword(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UsersController struct {
	router      *gin.Engine
	authService AuthService
	userService UserService
	mw          *middlewares.Middlewares
}

func NewUsersController(engine *gin.Engine, authService AuthService, userService UserService, mw *middlewares.Middlewares) *UsersController {
	return &UsersController{router: engine, authService: authService, userService: userService, mw: mw}
}

func (ctrl *UsersController) RegisterRoutes() {
	basePath := ctrl.router.Group(viper.GetString(config.ApiBasePath))
	authRoutes := basePath.Group("/auth")
	{
		authRoutes.POST("/register", ctrl.Register)
		authRoutes.POST("/login", ctrl.LogIn)
		authRoutes.POST("/refresh", ctrl.RefreshTokens)
		authRoutes.POST("/logout", ctrl.mw.AuthMiddleware(), ctrl.LogOut)
	}
	userRoutes := basePath.Group("/users")
	{
		authedUserRoutes := userRoutes.Use(ctrl.mw.AuthMiddleware())
		{
			authedUserRoutes.GET("/me", ctrl.GetCurrentUser)
			authedUserRoutes.PATCH("/me", ctrl.UpdateCurrentUser)
			authedUserRoutes.DELETE("/me", ctrl.DeleteCurrentUser)
			authedUserRoutes.PATCH("/update-password", ctrl.UpdateCurrentPassword)

			authedUserRoutes.GET("/:uuid", ctrl.GetUserByID)
		}
	}
}

func (ctrl *UsersController) Register(ctx *gin.Context) {
	var req models.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	user, err := ctrl.userService.Register(ctx, req)
	if err != nil {
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

func (ctrl *UsersController) LogIn(ctx *gin.Context) {
	var req models.LogInUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
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

func (ctrl *UsersController) RefreshTokens(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
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

func (ctrl *UsersController) LogOut(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
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

func (ctrl *UsersController) GetCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, user.ToPrivateResponse())
}

func (ctrl *UsersController) GetUserByID(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, user.ToPublicResponse())
}

func (ctrl *UsersController) UpdateCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.userService.UpdateUserByID(ctx, userID, req); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, user.ToPrivateResponse())
}

func (ctrl *UsersController) UpdateCurrentPassword(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	var req models.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.userService.ChangePassword(ctx, userID, req); err != nil {
		if _, valid := errors.AsType[svcErr.ValidationError](err); valid {
			apiModels.Error(ctx, http.StatusBadRequest, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "password changed"})
}

func (ctrl *UsersController) DeleteCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	var req models.DeleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.userService.VerifyPassword(ctx, userID, req.CurrentPassword); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.userService.Delete(ctx, userID); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(200, apiModels.APIResponse{Message: "account deleted"})
}
