package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
	"wishlist/internal/utils/str"
)

type ListStorage interface {
	CreateList(ctx context.Context, list models.List) error
	GetListByID(ctx context.Context, id uuid.UUID) (models.List, error)
	GetListBySharedLink(ctx context.Context, token string) (models.List, error)
	GetListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	UpdateListByID(ctx context.Context, id uuid.UUID, req models.UpdateListRequest) error
	RotateSharedLink(ctx context.Context, id uuid.UUID, token string) error
	DeleteListByID(ctx context.Context, id uuid.UUID) error
}

type WishStorage interface {
	GetWishesByListID(ctx context.Context, listID uuid.UUID) ([]models.Wish, error)
}

type ListServiceImpl struct {
	listStorage ListStorage
	wishStorage WishStorage
}

func NewListService(ls ListStorage, ws WishStorage) *ListServiceImpl {
	return &ListServiceImpl{listStorage: ls, wishStorage: ws}
}

func (svc *ListServiceImpl) CreateList(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error) {
	token, err := str.GenerateRandomString(16)
	if err != nil {
		return models.List{}, fmt.Errorf("failed to generate share token: %w", err)
	}

	list := models.List{
		ID:         uuid.New(),
		UserID:     userID,
		Title:      req.Title,
		Notes:      req.Notes,
		IsPublic:   true, // default
		ShareToken: token,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err = svc.listStorage.CreateList(ctx, list); err != nil {
		return models.List{}, err
	}

	return list, nil
}

func (svc *ListServiceImpl) GetListByID(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, error) {
	list, err := svc.listStorage.GetListByID(ctx, listID)
	if err != nil {
		return models.List{}, err
	}

	if list.UserID != requestedByUserID && !list.IsPublic {
		return models.List{}, svcErr.ForbiddenError{Message: "this wishlist is private"}
	}

	return list, nil
}

func (svc *ListServiceImpl) GetListBySharedLink(ctx context.Context, token string) (models.List, error) {
	return svc.listStorage.GetListBySharedLink(ctx, token)
}

func (svc *ListServiceImpl) GetListWithWishes(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, []models.Wish, error) {
	list, err := svc.GetListByID(ctx, listID, requestedByUserID)
	if err != nil {
		return models.List{}, nil, err
	}

	wishes, err := svc.wishStorage.GetWishesByListID(ctx, listID)
	if err != nil {
		return models.List{}, nil, err
	}

	return list, wishes, nil
}

func (svc *ListServiceImpl) GetListWithWishesBySharedLink(ctx context.Context, token string) (models.List, []models.Wish, error) {
	list, err := svc.listStorage.GetListBySharedLink(ctx, token)
	if err != nil {
		return models.List{}, nil, err
	}

	wishes, err := svc.wishStorage.GetWishesByListID(ctx, list.ID)
	if err != nil {
		return models.List{}, nil, err
	}

	return list, wishes, nil
}

func (svc *ListServiceImpl) GetCurrentUserLists(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return svc.listStorage.GetListsByUserID(ctx, userID)
}

func (svc *ListServiceImpl) GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return svc.listStorage.GetPublicListsByUserID(ctx, userID)
}

func (svc *ListServiceImpl) UpdateList(ctx context.Context, listID, userID uuid.UUID, req models.UpdateListRequest) error {
	list, err := svc.listStorage.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list.UserID != userID {
		return svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	return svc.listStorage.UpdateListByID(ctx, listID, req)
}

func (svc *ListServiceImpl) RotateSharedLink(ctx context.Context, listID, userID uuid.UUID) (string, error) {
	list, err := svc.listStorage.GetListByID(ctx, listID)
	if err != nil {
		return "", err
	}

	if list.UserID != userID {
		return "", svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	token, err := str.GenerateRandomString(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate share token: %w", err)
	}

	if err = svc.listStorage.RotateSharedLink(ctx, listID, token); err != nil {
		return "", err
	}

	return token, nil
}

func (svc *ListServiceImpl) DeleteList(ctx context.Context, listID, userID uuid.UUID) error {
	list, err := svc.listStorage.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list.UserID != userID {
		return svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	return svc.listStorage.DeleteListByID(ctx, listID)
}
