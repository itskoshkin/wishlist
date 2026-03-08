package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"wishlist/internal/models"
	svcErr "wishlist/internal/services/errors"
)

type listStorageMock struct {
	createErr error
	getErr    error
	updateErr error
	rotateErr error
	deleteErr error

	listToReturn  models.List
	listsToReturn []models.List

	createdList models.List
	rotatedID   uuid.UUID
	rotatedWith string
}

func (m *listStorageMock) CreateList(ctx context.Context, list models.List) error {
	m.createdList = list
	return m.createErr
}

func (m *listStorageMock) GetListByID(ctx context.Context, id uuid.UUID) (models.List, error) {
	if m.getErr != nil {
		return models.List{}, m.getErr
	}
	return m.listToReturn, nil
}

func (m *listStorageMock) GetListBySharedLink(ctx context.Context, token string) (models.List, error) {
	if m.getErr != nil {
		return models.List{}, m.getErr
	}
	return m.listToReturn, nil
}

func (m *listStorageMock) GetListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return m.listsToReturn, nil
}

func (m *listStorageMock) GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	return m.listsToReturn, nil
}

func (m *listStorageMock) UpdateListByID(ctx context.Context, id uuid.UUID, req models.UpdateListRequest) error {
	return m.updateErr
}

func (m *listStorageMock) RotateSharedLink(ctx context.Context, id uuid.UUID, token string) error {
	m.rotatedID = id
	m.rotatedWith = token
	return m.rotateErr
}

func (m *listStorageMock) DeleteListByID(ctx context.Context, id uuid.UUID) error {
	return m.deleteErr
}

type listWishStorageMock struct {
	wishes    []models.Wish
	wishesErr error
}

func (m *listWishStorageMock) CreateWish(ctx context.Context, wish models.Wish) error { return nil }

func (m *listWishStorageMock) GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error) {
	return models.Wish{}, nil
}

func (m *listWishStorageMock) GetWishesByListID(ctx context.Context, listID uuid.UUID) ([]models.Wish, error) {
	if m.wishesErr != nil {
		return nil, m.wishesErr
	}
	return m.wishes, nil
}

func (m *listWishStorageMock) UpdateWishByID(ctx context.Context, wishID uuid.UUID, req models.UpdateWishRequest) error {
	return nil
}

func (m *listWishStorageMock) ReserveWish(ctx context.Context, wishID, userID uuid.UUID) error {
	return nil
}

func (m *listWishStorageMock) ReleaseWish(ctx context.Context, wishID, userID uuid.UUID) error {
	return nil
}

func (m *listWishStorageMock) DeleteWishByID(ctx context.Context, wishID uuid.UUID) error { return nil }

func TestListService_CreateList_DefaultsAndToken(t *testing.T) {
	ls := &listStorageMock{}
	ws := &listWishStorageMock{}
	svc := NewListService(ls, ws)

	userID := uuid.New()
	title := "Birthday"
	notes := "for 2026"
	list, err := svc.CreateList(context.Background(), userID, models.CreateListRequest{
		Title: title,
		Notes: &notes,
	})
	if err != nil {
		t.Fatalf("CreateList() error = %v", err)
	}
	if !list.IsPublic {
		t.Fatal("CreateList() IsPublic = false, want true")
	}
	if len(list.ShareToken) != 32 {
		t.Fatalf("CreateList() ShareToken len = %d, want 32", len(list.ShareToken))
	}
	if ls.createdList.UserID != userID {
		t.Fatalf("CreateList() stored UserID = %s, want %s", ls.createdList.UserID, userID)
	}
}

func TestListService_GetListByID_PrivateForbidden(t *testing.T) {
	ownerID := uuid.New()
	requestedBy := uuid.New()
	ls := &listStorageMock{listToReturn: models.List{
		ID:       uuid.New(),
		UserID:   ownerID,
		IsPublic: false,
	}}
	svc := NewListService(ls, &listWishStorageMock{})

	_, err := svc.GetListByID(context.Background(), uuid.New(), requestedBy)
	if err == nil {
		t.Fatal("GetListByID() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("GetListByID() error = %T, want ForbiddenError", err)
	}
}

func TestListService_GetListWithWishes(t *testing.T) {
	listID := uuid.New()
	userID := uuid.New()
	wishes := []models.Wish{{ID: uuid.New(), ListID: listID}}
	ls := &listStorageMock{listToReturn: models.List{
		ID:       listID,
		UserID:   userID,
		IsPublic: true,
	}}
	ws := &listWishStorageMock{wishes: wishes}
	svc := NewListService(ls, ws)

	gotList, gotWishes, err := svc.GetListWithWishes(context.Background(), listID, userID)
	if err != nil {
		t.Fatalf("GetListWithWishes() error = %v", err)
	}
	if gotList.ID != listID {
		t.Fatalf("GetListWithWishes() list ID = %s, want %s", gotList.ID, listID)
	}
	if len(gotWishes) != 1 {
		t.Fatalf("GetListWithWishes() wishes len = %d, want 1", len(gotWishes))
	}
}

func TestListService_UpdateList_NotOwner(t *testing.T) {
	ownerID := uuid.New()
	callerID := uuid.New()
	ls := &listStorageMock{listToReturn: models.List{
		ID:     uuid.New(),
		UserID: ownerID,
	}}
	svc := NewListService(ls, &listWishStorageMock{})

	err := svc.UpdateList(context.Background(), uuid.New(), callerID, models.UpdateListRequest{})
	if err == nil {
		t.Fatal("UpdateList() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("UpdateList() error = %T, want ForbiddenError", err)
	}
}

func TestListService_RotateSharedLink(t *testing.T) {
	listID := uuid.New()
	userID := uuid.New()
	ls := &listStorageMock{listToReturn: models.List{
		ID:     listID,
		UserID: userID,
	}}
	svc := NewListService(ls, &listWishStorageMock{})

	token, err := svc.RotateSharedLink(context.Background(), listID, userID)
	if err != nil {
		t.Fatalf("RotateSharedLink() error = %v", err)
	}
	if len(token) != 32 {
		t.Fatalf("RotateSharedLink() token len = %d, want 32", len(token))
	}
	if ls.rotatedID != listID {
		t.Fatalf("RotateSharedLink() storage listID = %s, want %s", ls.rotatedID, listID)
	}
	if ls.rotatedWith != token {
		t.Fatalf("RotateSharedLink() storage token mismatch")
	}
}

func TestListService_GetListBySharedLink(t *testing.T) {
	expected := models.List{ID: uuid.New(), ShareToken: "12345678901234567890123456789012"}
	ls := &listStorageMock{listToReturn: expected}
	svc := NewListService(ls, &listWishStorageMock{})

	actual, err := svc.GetListBySharedLink(context.Background(), expected.ShareToken)
	if err != nil {
		t.Fatalf("GetListBySharedLink() error = %v", err)
	}
	if actual.ID != expected.ID {
		t.Fatalf("GetListBySharedLink() ID = %s, want %s", actual.ID, expected.ID)
	}
}

func TestListService_GetListWithWishesBySharedLink(t *testing.T) {
	list := models.List{ID: uuid.New(), ShareToken: "12345678901234567890123456789012"}
	wishes := []models.Wish{{ID: uuid.New(), ListID: list.ID}}
	ls := &listStorageMock{listToReturn: list}
	ws := &listWishStorageMock{wishes: wishes}
	svc := NewListService(ls, ws)

	gotList, gotWishes, err := svc.GetListWithWishesBySharedLink(context.Background(), list.ShareToken)
	if err != nil {
		t.Fatalf("GetListWithWishesBySharedLink() error = %v", err)
	}
	if gotList.ID != list.ID {
		t.Fatalf("list ID = %s, want %s", gotList.ID, list.ID)
	}
	if len(gotWishes) != 1 {
		t.Fatalf("wishes len = %d, want 1", len(gotWishes))
	}
}

func TestListService_GetCurrentAndPublicUserLists(t *testing.T) {
	userID := uuid.New()
	expected := []models.List{{ID: uuid.New(), UserID: userID}}
	ls := &listStorageMock{listsToReturn: expected}
	svc := NewListService(ls, &listWishStorageMock{})

	current, err := svc.GetCurrentUserLists(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCurrentUserLists() error = %v", err)
	}
	if len(current) != 1 || current[0].ID != expected[0].ID {
		t.Fatalf("GetCurrentUserLists() mismatch")
	}

	public, err := svc.GetPublicListsByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetPublicListsByUserID() error = %v", err)
	}
	if len(public) != 1 || public[0].ID != expected[0].ID {
		t.Fatalf("GetPublicListsByUserID() mismatch")
	}
}

func TestListService_DeleteList(t *testing.T) {
	listID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()
	ls := &listStorageMock{listToReturn: models.List{ID: listID, UserID: ownerID}}
	svc := NewListService(ls, &listWishStorageMock{})

	err := svc.DeleteList(context.Background(), listID, callerID)
	if err == nil {
		t.Fatal("DeleteList() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("DeleteList() error = %T, want ForbiddenError", err)
	}

	err = svc.DeleteList(context.Background(), listID, ownerID)
	if err != nil {
		t.Fatalf("DeleteList() owner error = %v", err)
	}
}

func TestListService_UpdateList_Success(t *testing.T) {
	listID := uuid.New()
	ownerID := uuid.New()
	ls := &listStorageMock{listToReturn: models.List{ID: listID, UserID: ownerID}}
	svc := NewListService(ls, &listWishStorageMock{})

	title := "Updated"
	err := svc.UpdateList(context.Background(), listID, ownerID, models.UpdateListRequest{Title: &title})
	if err != nil {
		t.Fatalf("UpdateList() error = %v", err)
	}
}

func TestListService_RotateSharedLink_NotOwner(t *testing.T) {
	listID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()
	ls := &listStorageMock{listToReturn: models.List{ID: listID, UserID: ownerID}}
	svc := NewListService(ls, &listWishStorageMock{})

	_, err := svc.RotateSharedLink(context.Background(), listID, callerID)
	if err == nil {
		t.Fatal("RotateSharedLink() error = nil, want forbidden")
	}
	var forbidden svcErr.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("RotateSharedLink() error = %T, want ForbiddenError", err)
	}
}
