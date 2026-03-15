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
	GetListBySharedLink(ctx context.Context, slug string) (models.List, error)
	GetListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error)
	UpdateListByID(ctx context.Context, id uuid.UUID, req models.UpdateListRequest) error
	RotateSharedLink(ctx context.Context, id uuid.UUID, slug string) error
	DeleteListByID(ctx context.Context, id uuid.UUID) error
}

type ListServiceImpl struct {
	lists  ListStorage
	wishes WishStorage
}

func NewListService(ls ListStorage, ws WishStorage) *ListServiceImpl {
	return &ListServiceImpl{lists: ls, wishes: ws}
}

func (svc *ListServiceImpl) CreateList(ctx context.Context, userID uuid.UUID, req models.CreateListRequest) (models.List, error) {
	slug, err := str.GenerateRandomString(16)
	if err != nil {
		return models.List{}, fmt.Errorf("failed to generate slug: %w", err)
	}

	list := models.List{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     req.Title,
		Notes:     req.Notes,
		IsPublic:  true, // default
		Slug:      slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = svc.lists.CreateList(ctx, list); err != nil {
		return models.List{}, err
	}

	return list, nil
}

func (svc *ListServiceImpl) GetListByID(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, error) {
	list, err := svc.lists.GetListByID(ctx, listID)
	if err != nil {
		return models.List{}, err
	}

	if list.UserID != requestedByUserID && !list.IsPublic {
		return models.List{}, svcErr.ForbiddenError{Message: "this wishlist is private"}
	}

	return list, nil
}

func (svc *ListServiceImpl) GetListBySharedLink(ctx context.Context, slug string) (models.List, error) {
	return svc.lists.GetListBySharedLink(ctx, slug)
}

func (svc *ListServiceImpl) GetListWithWishes(ctx context.Context, listID, requestedByUserID uuid.UUID) (models.List, []models.Wish, error) {
	list, err := svc.GetListByID(ctx, listID, requestedByUserID)
	if err != nil {
		return models.List{}, nil, err
	}

	wishes, err := svc.wishes.GetWishesByListID(ctx, listID)
	if err != nil {
		return models.List{}, nil, err
	}

	return list, wishes, nil
}

func (svc *ListServiceImpl) GetListWithWishesBySharedLink(ctx context.Context, slug string) (models.List, []models.Wish, error) {
	list, err := svc.lists.GetListBySharedLink(ctx, slug)
	if err != nil {
		return models.List{}, nil, err
	}

	wishes, err := svc.wishes.GetWishesByListID(ctx, list.ID)
	if err != nil {
		return models.List{}, nil, err
	}

	return list, wishes, nil
}

func (svc *ListServiceImpl) GetCurrentUserLists(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return svc.lists.GetListsByUserID(ctx, userID)
}

func (svc *ListServiceImpl) GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return svc.lists.GetPublicListsByUserID(ctx, userID)
}

func (svc *ListServiceImpl) UpdateList(ctx context.Context, listID, userID uuid.UUID, req models.UpdateListRequest) error {
	list, err := svc.lists.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list.UserID != userID {
		return svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	return svc.lists.UpdateListByID(ctx, listID, req)
}

func (svc *ListServiceImpl) RotateSharedLink(ctx context.Context, listID, userID uuid.UUID) (string, error) {
	list, err := svc.lists.GetListByID(ctx, listID)
	if err != nil {
		return "", err
	}

	if list.UserID != userID {
		return "", svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	slug, err := str.GenerateRandomString(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate slug: %w", err)
	}

	if err = svc.lists.RotateSharedLink(ctx, listID, slug); err != nil {
		return "", err
	}

	return slug, nil
}

func (svc *ListServiceImpl) DeleteList(ctx context.Context, listID, userID uuid.UUID) error {
	list, err := svc.lists.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	if list.UserID != userID {
		return svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	return svc.lists.DeleteListByID(ctx, listID)
}
