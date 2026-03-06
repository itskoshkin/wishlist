package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
	"wishlist/internal/storage"
	"wishlist/internal/utils/str"
)

type EmailService interface {
	SendPasswordResetLetter(ctx context.Context, to, token string) error
	SendEmailVerificationLetter(ctx context.Context, to, token string) error
}

type UserStorage interface {
	CreateUser(ctx context.Context, user models.User) error
	SetUserEmailAsVerified(ctx context.Context, id uuid.UUID) error
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error
	RemoveUserAvatar(ctx context.Context, id uuid.UUID) error
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

type AvatarStorage interface {
	GetBaseURL() string
	//UploadObject(ctx context.Context, objectName, filePath, contentType string) error
	UploadObject(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	GetObjectURL(objectName string) string
	DeleteObject(ctx context.Context, objectName string) error
}

type Logger interface {
	Error(format string, v ...any)
}

type UserServiceImpl struct {
	email   EmailService
	tokens  TokenStorage
	storage UserStorage
	s3      AvatarStorage
	log     Logger //MARK: Unsure if it is a good idea, but definitely better than putting logger from controller
}

func NewUserService(es EmailService, us UserStorage, ts TokenStorage, ms AvatarStorage, l Logger) *UserServiceImpl {
	return &UserServiceImpl{email: es, tokens: ts, storage: us, s3: ms, log: l}
}

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

	if req.Email != nil {
		var token string
		token, err = str.GenerateRandomString()
		if err != nil {
			return models.User{}, fmt.Errorf("failed to generate token: %w", err)
		}

		if err = svc.tokens.SaveEmailVerificationToken(ctx, token, user.ID.String()); err != nil {
			svc.log.Error("failed to save email verification token for user '%s': %v", user.ID, err)
		}

		if err = svc.email.SendEmailVerificationLetter(ctx, *req.Email, token); err != nil {
			svc.log.Error("failed to send verification email for user '%s': %v", user.ID, err)
		}
	}

	return user, nil
}

func (svc *UserServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	userIDStr, err := svc.tokens.GetEmailVerificationToken(ctx, token)
	if err != nil {
		return svcErr.ValidationError{Message: "invalid or expired verification token"}
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	if err = svc.storage.SetUserEmailAsVerified(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	if err = svc.tokens.DeleteEmailVerificationToken(ctx, token); err != nil {
		svc.log.Error("failed to delete email verification token for user '%s': %v", userID, err)
	}

	return nil
}

func (svc *UserServiceImpl) LogIn(ctx context.Context, req models.LogInUserRequest) (models.User, error) {
	user, err := svc.storage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return models.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return models.User{}, fmt.Errorf("invalid password: %w", err)
	}

	//if user.Email != nil && !user.EmailVerified {
	//	return models.User{}, svcErr.ValidationError{Message: "email not verified"}
	//}

	return user, nil
}

func (svc *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	return svc.storage.GetUserByID(ctx, id)
}

func (svc *UserServiceImpl) UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
	return svc.storage.UpdateUserByID(ctx, id, req)
}

func (svc *UserServiceImpl) UpdateAvatar(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) error {
	objectName := fmt.Sprintf(storage.AvatarPrefix, id, uuid.NewString())

	if err := svc.s3.UploadObject(ctx, objectName, reader, size, contentType); err != nil {
		return fmt.Errorf("failed to upload avatar: %w", err)
	}

	user, err := svc.storage.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	avatarURL := svc.s3.GetObjectURL(objectName)
	if err = svc.storage.UpdateUserByID(ctx, id, models.UpdateUserRequest{Avatar: &avatarURL}); err != nil {
		return fmt.Errorf("failed to save avatar URL: %w", err)
	}

	if user.Avatar != nil {
		objectName = strings.TrimPrefix(*user.Avatar, svc.s3.GetBaseURL()+"/")
		if err = svc.s3.DeleteObject(ctx, objectName); err != nil {
			svc.log.Error("failed to delete previous avatar for user with ID '%s': %v", user.ID.String(), err)
		}
	}

	return nil
}

func (svc *UserServiceImpl) DeleteAvatar(ctx context.Context, id uuid.UUID) error {
	user, err := svc.storage.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Avatar == nil {
		return nil // Nothing to delete
	}

	objectName := strings.TrimPrefix(*user.Avatar, svc.s3.GetBaseURL()+"/")
	if err = svc.s3.DeleteObject(ctx, objectName); err != nil {
		svc.log.Error("failed to delete avatar for user with ID '%s': %v", id, err)
	}

	return svc.storage.RemoveUserAvatar(ctx, id)
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

func (svc *UserServiceImpl) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := svc.storage.GetUserByEmail(ctx, email)
	if err != nil {
		return nil // No user? Keep it silent
	}
	if user.Email == nil {
		return nil // Also sealed lips
	}

	token, err := str.GenerateRandomString()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	if err = svc.tokens.SavePasswordResetToken(ctx, token, user.ID.String()); err != nil {
		return fmt.Errorf("failed to save password reset request: %w", err)
	}

	if err = svc.email.SendPasswordResetLetter(ctx, *user.Email, token); err != nil {
		return fmt.Errorf("failed to send password reset link: %w", err)
	}

	return nil
}

func (svc *UserServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) error {
	userIdString, err := svc.tokens.GetPasswordResetToken(ctx, token)
	if err != nil {
		return svcErr.ValidationError{Message: "invalid or expired password reset token"}
	}

	userID, err := uuid.Parse(userIdString)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newPass := string(hash)
	if err = svc.storage.UpdateUserByID(ctx, userID, models.UpdateUserRequest{Password: &newPass}); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err = svc.tokens.DeletePasswordResetToken(ctx, token); err != nil {
		//logger.ErrorWithID(ctx, fmt.Sprintf("failed to delete password reset token for user with ID '%s': %v", userID.String(), err)) //MARK: Bad idea
		svc.log.Error("failed to delete password reset token for user with ID '%s': %v", userID, err)
	}

	return nil
}

func (svc *UserServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return svc.storage.DeleteUserByID(ctx, id)
}
