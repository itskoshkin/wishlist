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
		authedListRoutes := listRoutes.Use(ctrl.mw.AuthMiddleware())
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
		authedUserRoutes := userRoutes.Use(ctrl.mw.AuthMiddleware())
		{
			authedUserRoutes.GET("/:user_id/lists", ctrl.GetPublicListsByUserID)
		}
	}
}

func (ctrl *ListsController) CreateList(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
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

func (ctrl *ListsController) GetListByID(ctx *gin.Context) {
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

func (ctrl *ListsController) GetCurrentUserLists(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
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

func (ctrl *ListsController) GetPublicListsByUserID(ctx *gin.Context) {
	currentUserID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "user ID not found in context")
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

func (ctrl *ListsController) UpdateList(ctx *gin.Context) {
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

func (ctrl *ListsController) RotateSharedLink(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, gin.H{"share_token": token})
}

func (ctrl *ListsController) DeleteList(ctx *gin.Context) {
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
