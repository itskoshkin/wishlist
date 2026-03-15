package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"wishlist/internal/models"
	svcErr "wishlist/internal/services/errors"
)

type wishSvcWishStorageMock struct {
	createErr  error
	getErr     error
	updateErr  error
	reserveErr error
	releaseErr error
	deleteErr  error

	wishToReturn models.Wish
	deletedID    uuid.UUID
	reservedID   uuid.UUID
	reservedBy   uuid.UUID
}

func (m *wishSvcWishStorageMock) CreateWish(ctx context.Context, wish models.Wish) error {
	return m.createErr
}

func (m *wishSvcWishStorageMock) GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error) {
	if m.getErr != nil {
		return models.Wish{}, m.getErr
	}
	return m.wishToReturn, nil
}

func (m *wishSvcWishStorageMock) GetWishesByListID(ctx context.Context, listID uuid.UUID) ([]models.Wish, error) {
	return nil, nil
}

func (m *wishSvcWishStorageMock) UpdateWishByID(ctx context.Context, wishID uuid.UUID, req models.UpdateWishRequest) error {
	return m.updateErr
}

func (m *wishSvcWishStorageMock) ReserveWish(ctx context.Context, wishID, userID uuid.UUID) error {
	m.reservedID = wishID
	m.reservedBy = userID
	return m.reserveErr
}

func (m *wishSvcWishStorageMock) ReleaseWish(ctx context.Context, wishID, userID uuid.UUID) error {
	return m.releaseErr
}

func (m *wishSvcWishStorageMock) DeleteWishByID(ctx context.Context, wishID uuid.UUID) error {
	m.deletedID = wishID
	return m.deleteErr
}

type wishSvcListStorageMock struct {
	getErr error
	list   models.List
}

func (m *wishSvcListStorageMock) CreateList(ctx context.Context, list models.List) error {
	return nil
}

func (m *wishSvcListStorageMock) GetListByID(ctx context.Context, id uuid.UUID) (models.List, error) {
	if m.getErr != nil {
		return models.List{}, m.getErr
	}
	return m.list, nil
}

func (m *wishSvcListStorageMock) GetListBySharedLink(ctx context.Context, token string) (models.List, error) {
	return models.List{}, nil
}

func (m *wishSvcListStorageMock) GetListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return nil, nil
}

func (m *wishSvcListStorageMock) GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return nil, nil
}

func (m *wishSvcListStorageMock) UpdateListByID(ctx context.Context, id uuid.UUID, req models.UpdateListRequest) error {
	return nil
}

func (m *wishSvcListStorageMock) RotateSharedLink(ctx context.Context, id uuid.UUID, token string) error {
	return nil
}

func (m *wishSvcListStorageMock) DeleteListByID(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestWishService_CreateWish_NotOwner(t *testing.T) {
	listID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()
	wishStorage := &wishSvcWishStorageMock{}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	_, err := svc.CreateWish(context.Background(), listID, callerID, models.CreateWishRequest{Title: "PS5"})
	if err == nil {
		t.Fatal("CreateWish() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("CreateWish() error = %T, want ForbiddenError", err)
	}
}

func TestWishService_UpdateWish_ListMismatch(t *testing.T) {
	givenListID := uuid.New()
	actualListID := uuid.New()
	ownerID := uuid.New()
	wishID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: actualListID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: actualListID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	err := svc.UpdateWish(context.Background(), givenListID, wishID, ownerID, models.UpdateWishRequest{})
	if err == nil {
		t.Fatal("UpdateWish() error = nil, want validation error")
	}
	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("UpdateWish() error = %T, want ValidationError", err)
	}
}

func TestWishService_UpdateWish_NotOwner(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	err := svc.UpdateWish(context.Background(), listID, wishID, callerID, models.UpdateWishRequest{})
	if err == nil {
		t.Fatal("UpdateWish() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("UpdateWish() error = %T, want ForbiddenError", err)
	}
}

func TestWishService_ReserveWish_OwnWish(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	err := svc.ReserveWish(context.Background(), listID, wishID, ownerID)
	if err == nil {
		t.Fatal("ReserveWish() error = nil, want validation error")
	}
	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ReserveWish() error = %T, want ValidationError", err)
	}
}

func TestWishService_ReserveWish_Success(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	if err := svc.ReserveWish(context.Background(), listID, wishID, callerID); err != nil {
		t.Fatalf("ReserveWish() error = %v", err)
	}
	if wishStorage.reservedID != wishID || wishStorage.reservedBy != callerID {
		t.Fatalf("ReserveWish() forwarded wrong params")
	}
}

func TestWishService_ReserveWish_AlreadyReserved(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
		reserveErr:   errors.New("failed to reserve wish with ID 'x': already reserved or not found"),
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	err := svc.ReserveWish(context.Background(), listID, wishID, callerID)
	if err == nil {
		t.Fatal("ReserveWish() error = nil, want validation error")
	}

	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ReserveWish() error = %T, want ValidationError", err)
	}
	if validation.Message != "wish is already reserved" {
		t.Fatalf("validation message = %q, want %q", validation.Message, "wish is already reserved")
	}
}

func TestWishService_DeleteWish_Success(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	if err := svc.DeleteWish(context.Background(), listID, wishID, ownerID); err != nil {
		t.Fatalf("DeleteWish() error = %v", err)
	}
	if wishStorage.deletedID != wishID {
		t.Fatalf("DeleteWish() deleted ID = %s, want %s", wishStorage.deletedID, wishID)
	}
}

func TestWishService_CreateWish_Success(t *testing.T) {
	listID := uuid.New()
	ownerID := uuid.New()
	wishStorage := &wishSvcWishStorageMock{}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	price := int64(5000)
	currency := "RUB"
	wish, err := svc.CreateWish(context.Background(), listID, ownerID, models.CreateWishRequest{
		Title:    "Headphones",
		Price:    &price,
		Currency: &currency,
	})
	if err != nil {
		t.Fatalf("CreateWish() error = %v", err)
	}
	if wish.ID == uuid.Nil || wish.ListID != listID || wish.Title != "Headphones" {
		t.Fatalf("CreateWish() returned invalid wish: %+v", wish)
	}
}

func TestWishService_GetWishByID(t *testing.T) {
	wishID := uuid.New()
	expected := models.Wish{ID: wishID, Title: "Keyboard"}
	wishStorage := &wishSvcWishStorageMock{wishToReturn: expected}
	svc := NewWishService(wishStorage, &wishSvcListStorageMock{}, nil)

	actual, err := svc.GetWishByID(context.Background(), wishID)
	if err != nil {
		t.Fatalf("GetWishByID() error = %v", err)
	}
	if actual.ID != expected.ID {
		t.Fatalf("GetWishByID() ID = %s, want %s", actual.ID, expected.ID)
	}
}

func TestWishService_UpdateWish_Success(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	if err := svc.UpdateWish(context.Background(), listID, wishID, ownerID, models.UpdateWishRequest{Title: ptr("New Title")}); err != nil {
		t.Fatalf("UpdateWish() error = %v", err)
	}
}

func TestWishService_ReleaseWish(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	userID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	svc := NewWishService(wishStorage, &wishSvcListStorageMock{}, nil)

	err := svc.ReleaseWish(context.Background(), uuid.New(), wishID, userID)
	if err == nil {
		t.Fatal("ReleaseWish() error = nil, want validation error")
	}
	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ReleaseWish() error = %T, want ValidationError", err)
	}

	if err = svc.ReleaseWish(context.Background(), listID, wishID, userID); err != nil {
		t.Fatalf("ReleaseWish() success error = %v", err)
	}
}

func TestWishService_ReleaseWish_NotReservedByUser(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	userID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
		releaseErr:   errors.New("failed to release wish with ID 'x': not reserved by you or not found"),
	}
	svc := NewWishService(wishStorage, &wishSvcListStorageMock{}, nil)

	err := svc.ReleaseWish(context.Background(), listID, wishID, userID)
	if err == nil {
		t.Fatal("ReleaseWish() error = nil, want validation error")
	}

	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ReleaseWish() error = %T, want ValidationError", err)
	}
	if validation.Message != "wish is not reserved by you" {
		t.Fatalf("validation message = %q, want %q", validation.Message, "wish is not reserved by you")
	}
}

func TestWishService_DeleteWish_NotOwner(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()

	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	err := svc.DeleteWish(context.Background(), listID, wishID, callerID)
	if err == nil {
		t.Fatal("DeleteWish() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("DeleteWish() error = %T, want ForbiddenError", err)
	}
}

func TestWishService_CreateWish_StorageError(t *testing.T) {
	listID := uuid.New()
	ownerID := uuid.New()
	wishStorage := &wishSvcWishStorageMock{createErr: errors.New("db error")}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	_, err := svc.CreateWish(context.Background(), listID, ownerID, models.CreateWishRequest{Title: "PS5"})
	if err == nil {
		t.Fatal("CreateWish() error = nil, want storage error")
	}
}

func TestWishService_UpdateWish_StorageError(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	ownerID := uuid.New()
	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: listID},
		updateErr:    errors.New("db error"),
	}
	listStorage := &wishSvcListStorageMock{list: models.List{ID: listID, UserID: ownerID}}
	svc := NewWishService(wishStorage, listStorage, nil)

	err := svc.UpdateWish(context.Background(), listID, wishID, ownerID, models.UpdateWishRequest{})
	if err == nil {
		t.Fatal("UpdateWish() error = nil, want storage error")
	}
}

func TestWishService_ReserveWish_ListMismatch(t *testing.T) {
	listID := uuid.New()
	wishID := uuid.New()
	userID := uuid.New()
	wishStorage := &wishSvcWishStorageMock{
		wishToReturn: models.Wish{ID: wishID, ListID: uuid.New()},
	}
	svc := NewWishService(wishStorage, &wishSvcListStorageMock{}, nil)

	err := svc.ReserveWish(context.Background(), listID, wishID, userID)
	if err == nil {
		t.Fatal("ReserveWish() error = nil, want validation error")
	}
}

func ptr(v string) *string {
	return &v
}
