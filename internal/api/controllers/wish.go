package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"wishlist/internal/api/errors"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/models"
	"wishlist/internal/services/errors"
)

type WishService interface {
	CreateWish(ctx context.Context, listID, userID uuid.UUID, req models.CreateWishRequest) (models.Wish, error)
	GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error)
	UpdateWish(ctx context.Context, listID, wishID, userID uuid.UUID, req models.UpdateWishRequest) error
	ReserveWish(ctx context.Context, listID, wishID, userID uuid.UUID) error
	ReleaseWish(ctx context.Context, listID, wishID, userID uuid.UUID) error
	DeleteWish(ctx context.Context, listID, wishID, userID uuid.UUID) error
}

type WishesController struct {
	router      *gin.Engine
	mw          *middlewares.Middlewares
	wishService WishService
}

func NewWishesController(e *gin.Engine, mw *middlewares.Middlewares, ws WishService) *WishesController {
	return &WishesController{router: e, mw: mw, wishService: ws}
}

func (ctrl *WishesController) RegisterRoutes() {
	basePath := ctrl.router.Group(viper.GetString(config.ApiBasePath))
	listRoutes := basePath.Group("/lists")
	{
		authedListRoutes := listRoutes.Use(ctrl.mw.AuthMiddleware())
		{
			authedListRoutes.POST("/:list_id/wishes", ctrl.CreateWish)
			authedListRoutes.PATCH("/:list_id/wishes/:wish_id", ctrl.UpdateWish)
			authedListRoutes.POST("/:list_id/wishes/:wish_id/reserve", ctrl.ReserveWish)
			authedListRoutes.DELETE("/:list_id/wishes/:wish_id/reserve", ctrl.ReleaseWish)
			authedListRoutes.DELETE("/:list_id/wishes/:wish_id", ctrl.DeleteWish)
		}
	}
}

// CreateWish GoDoc
// @Summary Create wish
// @Description Create wish in wishlist
// @Tags wishes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Param request body models.CreateWishRequest true "Wish data"
// @Success 201 {object} models.WishResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id}/wishes [post]
func (ctrl *WishesController) CreateWish(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	var req models.CreateWishRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	wish, err := ctrl.wishService.CreateWish(ctx, listID, userID, req)
	if err != nil {
		var forbiddenError svcErr.ForbiddenError
		if errors.As(err, &forbiddenError) {
			apiModels.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, wish.ToOwnerResponse())
}

// UpdateWish GoDoc
// @Summary Update wish
// @Description Update wish fields
// @Tags wishes
// @Accept json
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Param wish_id path string true "Wish ID (UUID)"
// @Param request body models.UpdateWishRequest true "Update payload"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id}/wishes/{wish_id} [patch]
func (ctrl *WishesController) UpdateWish(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	wishID, err := uuid.Parse(ctx.Param("wish_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid wish ID")
		return
	}

	var req models.UpdateWishRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err = ctrl.wishService.UpdateWish(ctx, listID, wishID, userID, req); err != nil {
		var forbiddenError svcErr.ForbiddenError
		if errors.As(err, &forbiddenError) {
			apiModels.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ReserveWish GoDoc
// @Summary Reserve wish
// @Description Reserve wish for current user
// @Tags wishes
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Param wish_id path string true "Wish ID (UUID)"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id}/wishes/{wish_id}/reserve [post]
func (ctrl *WishesController) ReserveWish(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	wishID, err := uuid.Parse(ctx.Param("wish_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid wish ID")
		return
	}

	if err = ctrl.wishService.ReserveWish(ctx, listID, wishID, userID); err != nil {
		var validationError svcErr.ValidationError
		if errors.As(err, &validationError) {
			apiModels.Error(ctx, http.StatusBadRequest, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ReleaseWish GoDoc
// @Summary Release wish reservation
// @Description Release wish reservation by current user
// @Tags wishes
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Param wish_id path string true "Wish ID (UUID)"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id}/wishes/{wish_id}/reserve [delete]
func (ctrl *WishesController) ReleaseWish(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	wishID, err := uuid.Parse(ctx.Param("wish_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid wish ID")
		return
	}

	if err = ctrl.wishService.ReleaseWish(ctx, listID, wishID, userID); err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// DeleteWish GoDoc
// @Summary Delete wish
// @Description Delete wish from wishlist
// @Tags wishes
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Param wish_id path string true "Wish ID (UUID)"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id}/wishes/{wish_id} [delete]
func (ctrl *WishesController) DeleteWish(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	wishID, err := uuid.Parse(ctx.Param("wish_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid wish ID")
		return
	}

	if err = ctrl.wishService.DeleteWish(ctx, listID, wishID, userID); err != nil {
		var forbiddenError svcErr.ForbiddenError
		if errors.As(err, &forbiddenError) {
			apiModels.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}
