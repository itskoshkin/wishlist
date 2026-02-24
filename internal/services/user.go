package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

type UserServiceImpl struct {
	storage UserStorage
}

func NewUserService(storage UserStorage) *UserServiceImpl { return &UserServiceImpl{storage: storage} }

func (svc *UserServiceImpl) Register(ctx context.Context, req models.RegisterUserRequest) (models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := uuid.NewV7() // Sortable to make life easier for dear Postgres (faster INSERT and better index performance)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	user := models.User{
		ID:        id,
		Name:      req.Name,
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hash),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = svc.storage.CreateUser(ctx, user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (svc *UserServiceImpl) LogIn(ctx context.Context, req models.LogInUserRequest) (models.User, error) {
	user, err := svc.storage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return models.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return models.User{}, fmt.Errorf("invalid password: %w", err)
	}

	return user, nil
}

func (svc *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	return svc.storage.GetUserByID(ctx, id)
}

func (svc *UserServiceImpl) UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
	return svc.storage.UpdateUserByID(ctx, id, req)
}

func (svc *UserServiceImpl) VerifyPassword(ctx context.Context, id uuid.UUID, password string) error {
	user, err := svc.storage.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return svcErr.ValidationError{Message: "wrong password"}
	}

	return nil
}

func (svc *UserServiceImpl) ChangePassword(ctx context.Context, id uuid.UUID, req models.ChangePasswordRequest) error {
	user, err := svc.storage.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return svcErr.ValidationError{Message: "wrong current password"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newPass := string(hash)
	return svc.storage.UpdateUserByID(ctx, id, models.UpdateUserRequest{
		Password: &newPass,
	})
}

func (svc *UserServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return svc.storage.DeleteUserByID(ctx, id)
}
