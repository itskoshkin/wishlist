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

type ListService interface {
	CreateList(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error)
	GetListByID(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, error)
	GetListBySharedLink(ctx context.Context, token string) (models.List, error)
	GetListWithWishes(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, []models.Wish, error)
	GetListWithWishesBySharedLink(ctx context.Context, token string) (models.List, []models.Wish, error)
	GetCurrentUserLists(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	UpdateList(ctx context.Context, listID, userID uuid.UUID, req models.UpdateListRequest) error
	RotateSharedLink(ctx context.Context, listID, userID uuid.UUID) (string, error)
	DeleteList(ctx context.Context, listID, userID uuid.UUID) error
}

type ListsController struct {
	router      *gin.Engine
	mw          *middlewares.Middlewares
	listService ListService
}

func NewListsController(e *gin.Engine, mw *middlewares.Middlewares, ls ListService) *ListsController {
	return &ListsController{router: e, mw: mw, listService: ls}
}

func (ctrl *ListsController) RegisterRoutes() {
	basePath := ctrl.router.Group(viper.GetString(config.ApiBasePath))
	listRoutes := basePath.Group("/lists")
	{
		authedListRoutes := listRoutes.Group("").Use(ctrl.mw.AuthMiddleware())
		{
			authedListRoutes.POST("", ctrl.CreateList)
			authedListRoutes.GET("", ctrl.GetCurrentUserLists)
			authedListRoutes.GET("/:list_id", ctrl.GetListByID)
			authedListRoutes.PATCH("/:list_id", ctrl.UpdateList)
			authedListRoutes.POST("/:list_id/rotate-share-link", ctrl.RotateSharedLink)
			authedListRoutes.DELETE("/:list_id", ctrl.DeleteList)
		}
		listRoutes.GET("/shared/:slug", ctrl.mw.OptionalAuthMiddleware(), ctrl.GetListBySharedLink)
	}
	userRoutes := basePath.Group("/users")
	{
		authedUserRoutes := userRoutes.Group("").Use(ctrl.mw.AuthMiddleware())
		{
			authedUserRoutes.GET("/:user_id/lists", ctrl.GetPublicListsByUserID)
		}
	}
}

// CreateList GoDoc
// @Summary Create wishlist
// @Description Create a new wishlist for current user
// @Tags lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateListRequest true "Wishlist data"
// @Success 201 {object} models.ListResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists [post]
func (ctrl *ListsController) CreateList(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	list, err := ctrl.listService.CreateList(ctx, userID, req)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, list.ToOwnerResponse())
}

// GetListByID GoDoc
// @Summary Get wishlist by ID
// @Description Get wishlist with wishes by list ID
// @Tags lists
// @Produce json
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Success 200 {object} models.ListResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id} [get]
func (ctrl *ListsController) GetListByID(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	list, wishes, err := ctrl.listService.GetListWithWishes(ctx, listID, userID)
	if err != nil {
		var forbiddenError svcErr.ForbiddenError
		if errors.As(err, &forbiddenError) {
			apiModels.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	var response models.ListResponse
	if list.UserID == userID {
		response = list.ToOwnerResponse()
		wishResponses := make([]models.WishResponse, len(wishes))
		for i, wish := range wishes {
			wishResponses[i] = wish.ToOwnerResponse()
		}
		response.Wishes = wishResponses
	} else {
		response = list.ToViewerResponse(&userID)
		wishResponses := make([]models.WishResponse, len(wishes))
		for i, wish := range wishes {
			wishResponses[i] = wish.ToViewerResponse(&userID)
		}
		response.Wishes = wishResponses
	}

	ctx.JSON(http.StatusOK, response)
}

// GetListBySharedLink GoDoc
// @Summary Get wishlist by shared link
// @Description Get wishlist with wishes by shared slug
// @Tags lists
// @Produce json
// @Param slug path string true "Shared slug (32 chars)"
// @Success 200 {object} models.ListResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 404 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/shared/{slug} [get]
func (ctrl *ListsController) GetListBySharedLink(ctx *gin.Context) {
	slug := ctx.Param("slug")
	if slug == "" || len(slug) != 32 {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list URL")
		return
	}

	list, wishes, err := ctrl.listService.GetListWithWishesBySharedLink(ctx, slug)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	var userID *uuid.UUID
	if uid, ok := middlewares.GetUserID(ctx); ok {
		userID = &uid
	}

	var response models.ListResponse
	if userID != nil && list.UserID == *userID {
		response = list.ToOwnerResponse()
		wishResponses := make([]models.WishResponse, len(wishes))
		for i, wish := range wishes {
			wishResponses[i] = wish.ToOwnerResponse()
		}
		response.Wishes = wishResponses
	} else {
		response = list.ToViewerResponse(userID)
		wishResponses := make([]models.WishResponse, len(wishes))
		for i, wish := range wishes {
			wishResponses[i] = wish.ToViewerResponse(userID)
		}
		response.Wishes = wishResponses
	}

	ctx.JSON(http.StatusOK, response)
}

// GetCurrentUserLists GoDoc
// @Summary Get current user wishlists
// @Description Get all wishlists of current user
// @Tags lists
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ListResponse
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists [get]
func (ctrl *ListsController) GetCurrentUserLists(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	lists, err := ctrl.listService.GetCurrentUserLists(ctx, userID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	response := make([]models.ListResponse, len(lists))
	for i, list := range lists {
		response[i] = list.ToOwnerResponse()
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPublicListsByUserID GoDoc
// @Summary Get public wishlists by user ID
// @Description Get public wishlists of selected user
// @Tags lists
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {array} models.ListResponse
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /users/{user_id}/lists [get]
func (ctrl *ListsController) GetPublicListsByUserID(ctx *gin.Context) {
	currentUserID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid user ID")
		return
	}

	lists, err := ctrl.listService.GetPublicListsByUserID(ctx, userID)
	if err != nil {
		apiModels.InternalError(ctx, err.Error())
		return
	}

	response := make([]models.ListResponse, len(lists))
	for i, list := range lists {
		if list.UserID == currentUserID {
			response[i] = list.ToOwnerResponse()
		} else {
			response[i] = list.ToViewerResponse(&currentUserID)
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateList GoDoc
// @Summary Update wishlist
// @Description Update wishlist fields
// @Tags lists
// @Accept json
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Param request body models.UpdateListRequest true "Update payload"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id} [patch]
func (ctrl *ListsController) UpdateList(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	var req models.UpdateListRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err = ctrl.listService.UpdateList(ctx, listID, userID, req); err != nil {
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

// RotateSharedLink GoDoc
// @Summary Rotate shared link
// @Description Generate new shared token for wishlist
// @Tags lists
// @Produce json
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id}/rotate-share-link [post]
func (ctrl *ListsController) RotateSharedLink(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	token, err := ctrl.listService.RotateSharedLink(ctx, listID, userID)
	if err != nil {
		var forbiddenError svcErr.ForbiddenError
		if errors.As(err, &forbiddenError) {
			apiModels.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		apiModels.InternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"slug": token})
}

// DeleteList GoDoc
// @Summary Delete wishlist
// @Description Delete wishlist by ID
// @Tags lists
// @Security BearerAuth
// @Param list_id path string true "List ID (UUID)"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} apiModels.APIError
// @Failure 401 {object} apiModels.APIError
// @Failure 403 {object} apiModels.APIError
// @Failure 500 {object} apiModels.APIError
// @Router /lists/{list_id} [delete]
func (ctrl *ListsController) DeleteList(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := uuid.Parse(ctx.Param("list_id"))
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "invalid list ID")
		return
	}

	if err = ctrl.listService.DeleteList(ctx, listID, userID); err != nil {
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
