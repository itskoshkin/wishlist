package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
	"wishlist/internal/storage"
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
	s3        AvatarStorage
}

func NewWishService(ws WishStorage, wl ListStorage, s3 AvatarStorage) *WishServiceImpl {
	return &WishServiceImpl{wishes: ws, wishlists: wl, s3: s3}
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

func (svc *WishServiceImpl) UpdateWishImage(ctx context.Context, listID, wishID, userID uuid.UUID, reader io.Reader, size int64, contentType string) error {
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

	objectName := fmt.Sprintf(storage.WishImagePrefix, wishID, uuid.NewString())

	if err = svc.s3.UploadObject(ctx, objectName, reader, size, contentType); err != nil {
		return fmt.Errorf("failed to upload wish image: %w", err)
	}

	return svc.wishes.UpdateWishByID(ctx, wishID, models.UpdateWishRequest{Image: new(svc.s3.GetObjectURL(objectName))})
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

	if err = svc.wishes.ReserveWish(ctx, wishID, userID); err != nil {
		if _, ok := errors.AsType[svcErr.ValidationError](err); ok {
			return err
		}
		if strings.Contains(strings.ToLower(err.Error()), "already reserved or not found") {
			return svcErr.ValidationError{Message: "wish is already reserved"}
		}
		return err
	}

	return nil
}

func (svc *WishServiceImpl) ReleaseWish(ctx context.Context, listID, wishID, userID uuid.UUID) error {
	wish, err := svc.wishes.GetWishByID(ctx, wishID)
	if err != nil {
		return err
	}

	if wish.ListID != listID {
		return svcErr.ValidationError{Message: "wish does not belong to this list"}
	}

	if err = svc.wishes.ReleaseWish(ctx, wishID, userID); err != nil {
		if _, ok := errors.AsType[svcErr.ValidationError](err); ok {
			return err
		}
		if strings.Contains(strings.ToLower(err.Error()), "not reserved by you or not found") {
			return svcErr.ValidationError{Message: "wish is not reserved by you"}
		}
		return err
	}

	return nil
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
