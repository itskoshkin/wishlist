package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
)

type WishStorage interface {
	CreateWish(ctx context.Context, wish models.Wish) error
	GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error)
	GetWishesByListID(ctx context.Context, listID uuid.UUID) ([]models.Wish, error)
	UpdateWishByID(ctx context.Context, wishID uuid.UUID, req models.UpdateWishRequest) error
	ReserveWish(ctx context.Context, wishID, userID uuid.UUID) error
	ReleaseWish(ctx context.Context, wishID, userID uuid.UUID) error
	DeleteWishByID(ctx context.Context, wishID uuid.UUID) error
}

type WishServiceImpl struct {
	wishes    WishStorage
	wishlists ListStorage
}

func NewWishService(ws WishStorage, wl ListStorage) *WishServiceImpl {
	return &WishServiceImpl{wishes: ws, wishlists: wl}
}

func (svc *WishServiceImpl) CreateWish(ctx context.Context, listID, userID uuid.UUID, req models.CreateWishRequest) (models.Wish, error) {
	list, err := svc.wishlists.GetListByID(ctx, listID)
	if err != nil {
		return models.Wish{}, err
	}

	if list.UserID != userID {
		return models.Wish{}, svcErr.ForbiddenError{Message: "you are not the owner of this wishlist"}
	}

	wish := models.Wish{
		ID:        uuid.New(),
		ListID:    listID,
		Title:     req.Title,
		Notes:     req.Notes,
		Link:      req.Link,
		Price:     req.Price,
		Currency:  req.Currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = svc.wishes.CreateWish(ctx, wish); err != nil {
		return models.Wish{}, err
	}

	return wish, nil
}

func (svc *WishServiceImpl) GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error) {
	return svc.wishes.GetWishByID(ctx, wishID)
}

func (svc *WishServiceImpl) UpdateWish(ctx context.Context, listID, wishID, userID uuid.UUID, req models.UpdateWishRequest) error {
	wish, err := svc.wishes.GetWishByID(ctx, wishID)
	if err != nil {
		return err
	}

	if wish.ListID != listID {
		return svcErr.ValidationError{Message: "wish does not belong to this list"}
	}

	list, err := svc.wishlists.GetListByID(ctx, wish.ListID)
	if err != nil {
		return err
	}

	if list.UserID != userID {
		return svcErr.ForbiddenError{Message: "you are not the owner of this wish"}
	}

	return svc.wishes.UpdateWishByID(ctx, wishID, req)
}

func (svc *WishServiceImpl) ReserveWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	wish, err := svc.wishes.GetWishByID(ctx, wishID)
	if err != nil {
		return err
	}

	if wish.ListID != listID {
		return svcErr.ValidationError{Message: "wish does not belong to this list"}
	}

	list, err := svc.wishlists.GetListByID(ctx, wish.ListID)
	if err != nil {
		return err
	}

	if list.UserID == userID {
		return svcErr.ValidationError{Message: "you cannot reserve your own wish"}
	}

	return svc.wishes.ReserveWish(ctx, wishID, userID)
}

func (svc *WishServiceImpl) ReleaseWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	wish, err := svc.wishes.GetWishByID(ctx, wishID)
	if err != nil {
		return err
	}

	if wish.ListID != listID {
		return svcErr.ValidationError{Message: "wish does not belong to this list"}
	}

	return svc.wishes.ReleaseWish(ctx, wishID, userID)
}

func (svc *WishServiceImpl) DeleteWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	wish, err := svc.wishes.GetWishByID(ctx, wishID)
	if err != nil {
		return err
	}

	if wish.ListID != listID {
		return svcErr.ValidationError{Message: "wish does not belong to this list"}
	}

	list, err := svc.wishlists.GetListByID(ctx, wish.ListID)
	if err != nil {
		return err
	}

	if list.UserID != userID {
		return svcErr.ForbiddenError{Message: "you are not the owner of this wish"}
	}

	return svc.wishes.DeleteWishByID(ctx, wishID)
}
