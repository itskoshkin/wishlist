package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"wishlist/internal/models"
	svcErr "wishlist/internal/services/errors"
)

type userEmailServiceMock struct {
	verificationTo    string
	verificationToken string
	verificationCalls int
	verificationErr   error

	resetTo    string
	resetToken string
	resetCalls int
	resetErr   error
}

func (m *userEmailServiceMock) SendPasswordResetLetter(ctx context.Context, to, token string) error {
	m.resetTo = to
	m.resetToken = token
	m.resetCalls++
	return m.resetErr
}

func (m *userEmailServiceMock) SendEmailVerificationLetter(ctx context.Context, to, token string) error {
	m.verificationTo = to
	m.verificationToken = token
	m.verificationCalls++
	return m.verificationErr
}

type userTokenStorageMock struct {
	saveEmailTokenID string
	saveEmailUserID  string
	saveEmailCalls   int
	saveEmailErr     error

	getEmailValue string
	getEmailErr   error

	deleteEmailCalls int
	deleteEmailErr   error

	saveResetTokenID string
	saveResetUserID  string
	saveResetCalls   int
	saveResetErr     error

	getResetValue string
	getResetErr   error

	deleteResetCalls int
	deleteResetErr   error
}

func (m *userTokenStorageMock) SaveEmailVerificationToken(ctx context.Context, tokenID, userID string) error {
	m.saveEmailTokenID = tokenID
	m.saveEmailUserID = userID
	m.saveEmailCalls++
	return m.saveEmailErr
}

func (m *userTokenStorageMock) GetEmailVerificationToken(ctx context.Context, tokenID string) (string, error) {
	if m.getEmailErr != nil {
		return "", m.getEmailErr
	}
	return m.getEmailValue, nil
}

func (m *userTokenStorageMock) DeleteEmailVerificationToken(ctx context.Context, tokenID string) error {
	m.deleteEmailCalls++
	return m.deleteEmailErr
}

func (m *userTokenStorageMock) CheckIfAuthTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	return false, nil
}

func (m *userTokenStorageMock) RevokeAuthTokens(ctx context.Context, tokenID string, remainingTTL time.Duration) error {
	return nil
}

func (m *userTokenStorageMock) SavePasswordResetToken(ctx context.Context, tokenID string, userID string) error {
	m.saveResetTokenID = tokenID
	m.saveResetUserID = userID
	m.saveResetCalls++
	return m.saveResetErr
}

func (m *userTokenStorageMock) GetPasswordResetToken(ctx context.Context, tokenID string) (string, error) {
	if m.getResetErr != nil {
		return "", m.getResetErr
	}
	return m.getResetValue, nil
}

func (m *userTokenStorageMock) DeletePasswordResetToken(ctx context.Context, tokenID string) error {
	m.deleteResetCalls++
	return m.deleteResetErr
}

type userStorageServiceMock struct {
	createErr error
	updateErr error
	getErr    error
	deleteErr error

	userByID models.User

	userByUsername    models.User
	userByUsernameErr error

	userByEmail    models.User
	userByEmailErr error

	createdUser     models.User
	updatedUserID   uuid.UUID
	updatedUserReq  models.UpdateUserRequest
	setVerifiedUser uuid.UUID
	setVerifiedErr  error

	removeAvatarCalls int
	removeAvatarErr   error

	deletedUserID uuid.UUID
}

func (m *userStorageServiceMock) CreateUser(ctx context.Context, user models.User) error {
	m.createdUser = user
	return m.createErr
}

func (m *userStorageServiceMock) SetUserEmailAsVerified(ctx context.Context, id uuid.UUID) error {
	m.setVerifiedUser = id
	return m.setVerifiedErr
}

func (m *userStorageServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	if m.getErr != nil {
		return models.User{}, m.getErr
	}
	return m.userByID, nil
}

func (m *userStorageServiceMock) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	if m.userByUsernameErr != nil {
		return models.User{}, m.userByUsernameErr
	}
	return m.userByUsername, nil
}

func (m *userStorageServiceMock) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	if m.userByEmailErr != nil {
		return models.User{}, m.userByEmailErr
	}
	return m.userByEmail, nil
}

func (m *userStorageServiceMock) UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
	m.updatedUserID = id
	m.updatedUserReq = req
	return m.updateErr
}

func (m *userStorageServiceMock) RemoveUserAvatar(ctx context.Context, id uuid.UUID) error {
	m.removeAvatarCalls++
	return m.removeAvatarErr
}

func (m *userStorageServiceMock) DeleteUserByID(ctx context.Context, id uuid.UUID) error {
	m.deletedUserID = id
	return m.deleteErr
}

type userAvatarStorageMock struct {
	baseURL string

	uploadedObjectName string
	uploadedType       string
	uploadedSize       int64
	uploadErr          error

	newObjectURL string
	deletedObj   string
	deleteErr    error
}

func (m *userAvatarStorageMock) GetBaseURL() string { return m.baseURL }

func (m *userAvatarStorageMock) UploadObject(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	m.uploadedObjectName = objectName
	m.uploadedType = contentType
	m.uploadedSize = size
	return m.uploadErr
}

func (m *userAvatarStorageMock) GetObjectURL(objectName string) string {
	if m.newObjectURL != "" {
		return m.newObjectURL
	}
	return m.baseURL + "/" + objectName
}

func (m *userAvatarStorageMock) DeleteObject(ctx context.Context, objectName string) error {
	m.deletedObj = objectName
	return m.deleteErr
}

type userLoggerMock struct{ calls int }

func (m *userLoggerMock) Error(format string, v ...any) { m.calls++ }

func TestUserService_Register_WithoutEmail(t *testing.T) {
	st := &userStorageServiceMock{}
	tk := &userTokenStorageMock{}
	mailer := &userEmailServiceMock{}
	s3 := &userAvatarStorageMock{baseURL: "http://minio:9000/wishlist"}
	log := &userLoggerMock{}
	svc := NewUserService(mailer, st, tk, s3, log)

	user, err := svc.Register(context.Background(), models.RegisterUserRequest{
		Name:     "John",
		Username: "johnny",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.ID == uuid.Nil {
		t.Fatal("Register() user ID is nil")
	}
	if tk.saveEmailCalls != 0 {
		t.Fatalf("SaveEmailVerificationToken calls = %d, want 0", tk.saveEmailCalls)
	}
	if mailer.verificationCalls != 0 {
		t.Fatalf("SendEmailVerificationLetter calls = %d, want 0", mailer.verificationCalls)
	}
}

func TestUserService_Register_WithEmail_SendsVerification(t *testing.T) {
	email := "user@example.com"
	st := &userStorageServiceMock{}
	tk := &userTokenStorageMock{}
	mailer := &userEmailServiceMock{}
	s3 := &userAvatarStorageMock{baseURL: "http://minio:9000/wishlist"}
	log := &userLoggerMock{}
	svc := NewUserService(mailer, st, tk, s3, log)

	user, err := svc.Register(context.Background(), models.RegisterUserRequest{
		Name:     "John",
		Username: "johnny",
		Email:    &email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if tk.saveEmailCalls != 1 {
		t.Fatalf("SaveEmailVerificationToken calls = %d, want 1", tk.saveEmailCalls)
	}
	if len(tk.saveEmailTokenID) != 64 {
		t.Fatalf("verification token len = %d, want 64", len(tk.saveEmailTokenID))
	}
	if tk.saveEmailUserID != user.ID.String() {
		t.Fatalf("saved user ID = %s, want %s", tk.saveEmailUserID, user.ID)
	}
	if mailer.verificationCalls != 1 {
		t.Fatalf("SendEmailVerificationLetter calls = %d, want 1", mailer.verificationCalls)
	}
	if mailer.verificationTo != email {
		t.Fatalf("verification email to = %s, want %s", mailer.verificationTo, email)
	}
	if mailer.verificationToken != tk.saveEmailTokenID {
		t.Fatal("verification token in email doesn't match saved token")
	}
}

func TestUserService_VerifyEmail_InvalidToken(t *testing.T) {
	st := &userStorageServiceMock{}
	tk := &userTokenStorageMock{getEmailErr: errors.New("not found")}
	mailer := &userEmailServiceMock{}
	s3 := &userAvatarStorageMock{baseURL: "http://minio:9000/wishlist"}
	log := &userLoggerMock{}
	svc := NewUserService(mailer, st, tk, s3, log)

	err := svc.VerifyEmail(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("VerifyEmail() error = nil, want validation error")
	}
	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("VerifyEmail() error = %T, want ValidationError", err)
	}
}

func TestUserService_UpdateAvatar_RemovesPreviousAvatar(t *testing.T) {
	id := uuid.New()
	oldAvatar := "http://minio:9000/wishlist/avatars/old-user/old-file"
	st := &userStorageServiceMock{
		userByID: models.User{ID: id, Avatar: &oldAvatar},
	}
	tk := &userTokenStorageMock{}
	mailer := &userEmailServiceMock{}
	s3 := &userAvatarStorageMock{
		baseURL:      "http://minio:9000/wishlist",
		newObjectURL: "http://minio:9000/wishlist/avatars/new-user/new-file",
	}
	log := &userLoggerMock{}
	svc := NewUserService(mailer, st, tk, s3, log)

	err := svc.UpdateAvatar(context.Background(), id, strings.NewReader("x"), 1, "image/png")
	if err != nil {
		t.Fatalf("UpdateAvatar() error = %v", err)
	}
	if st.updatedUserID != id {
		t.Fatalf("UpdateAvatar() updated user ID = %s, want %s", st.updatedUserID, id)
	}
	if st.updatedUserReq.Avatar == nil || *st.updatedUserReq.Avatar != s3.newObjectURL {
		t.Fatalf("UpdateAvatar() avatar URL not saved")
	}
	if s3.deletedObj != "avatars/old-user/old-file" {
		t.Fatalf("UpdateAvatar() deleted object = %s, want avatars/old-user/old-file", s3.deletedObj)
	}
}

func TestUserService_VerifyEmail_Success(t *testing.T) {
	userID := uuid.New()
	st := &userStorageServiceMock{}
	tk := &userTokenStorageMock{getEmailValue: userID.String()}
	mailer := &userEmailServiceMock{}
	s3 := &userAvatarStorageMock{baseURL: "http://minio:9000/wishlist"}
	log := &userLoggerMock{}
	svc := NewUserService(mailer, st, tk, s3, log)

	if err := svc.VerifyEmail(context.Background(), "token"); err != nil {
		t.Fatalf("VerifyEmail() error = %v", err)
	}
	if st.setVerifiedUser != userID {
		t.Fatalf("verified user ID = %s, want %s", st.setVerifiedUser, userID)
	}
	if tk.deleteEmailCalls != 1 {
		t.Fatalf("DeleteEmailVerificationToken calls = %d, want 1", tk.deleteEmailCalls)
	}
}

func TestUserService_LogIn_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	userID := uuid.New()
	expected := models.User{ID: userID, Username: "johnny", Password: string(hash)}
	st := &userStorageServiceMock{userByUsername: expected}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})

	actual, err := svc.LogIn(context.Background(), models.LogInUserRequest{Username: "johnny", Password: "password123"})
	if err != nil {
		t.Fatalf("LogIn() error = %v", err)
	}
	if actual.ID != userID {
		t.Fatalf("LogIn() user ID = %s, want %s", actual.ID, userID)
	}
}

func TestUserService_LogIn_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	st := &userStorageServiceMock{
		userByUsername: models.User{ID: uuid.New(), Username: "johnny", Password: string(hash)},
	}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})

	err = nil
	_, err = svc.LogIn(context.Background(), models.LogInUserRequest{Username: "johnny", Password: "bad-pass"})
	if err == nil {
		t.Fatal("LogIn() error = nil, want invalid password")
	}
}

func TestUserService_GetAndUpdateByID(t *testing.T) {
	userID := uuid.New()
	expected := models.User{ID: userID, Username: "alice"}
	st := &userStorageServiceMock{userByID: expected}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})

	user, err := svc.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if user.ID != expected.ID {
		t.Fatalf("GetUserByID() ID = %s, want %s", user.ID, expected.ID)
	}

	name := "Alice Updated"
	if err = svc.UpdateUserByID(context.Background(), userID, models.UpdateUserRequest{Name: &name}); err != nil {
		t.Fatalf("UpdateUserByID() error = %v", err)
	}
	if st.updatedUserID != userID {
		t.Fatalf("updated user ID = %s, want %s", st.updatedUserID, userID)
	}
	if st.updatedUserReq.Name == nil || *st.updatedUserReq.Name != name {
		t.Fatal("updated name mismatch")
	}
}

func TestUserService_DeleteAvatar_NoAvatar(t *testing.T) {
	id := uuid.New()
	st := &userStorageServiceMock{userByID: models.User{ID: id}}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{baseURL: "http://minio"}, &userLoggerMock{})

	if err := svc.DeleteAvatar(context.Background(), id); err != nil {
		t.Fatalf("DeleteAvatar() error = %v", err)
	}
	if st.removeAvatarCalls != 0 {
		t.Fatalf("RemoveUserAvatar calls = %d, want 0", st.removeAvatarCalls)
	}
}

func TestUserService_DeleteAvatar_WithAvatar(t *testing.T) {
	id := uuid.New()
	avatar := "http://minio:9000/wishlist/avatars/user/file"
	st := &userStorageServiceMock{userByID: models.User{ID: id, Avatar: &avatar}}
	s3 := &userAvatarStorageMock{baseURL: "http://minio:9000/wishlist"}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, s3, &userLoggerMock{})

	if err := svc.DeleteAvatar(context.Background(), id); err != nil {
		t.Fatalf("DeleteAvatar() error = %v", err)
	}
	if s3.deletedObj != "avatars/user/file" {
		t.Fatalf("deleted object = %s, want avatars/user/file", s3.deletedObj)
	}
	if st.removeAvatarCalls != 1 {
		t.Fatalf("RemoveUserAvatar calls = %d, want 1", st.removeAvatarCalls)
	}
}

func TestUserService_VerifyPassword(t *testing.T) {
	id := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	st := &userStorageServiceMock{userByID: models.User{ID: id, Password: string(hash)}}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})

	if err = svc.VerifyPassword(context.Background(), id, "secret123"); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	err = svc.VerifyPassword(context.Background(), id, "bad")
	if err == nil {
		t.Fatal("VerifyPassword() error = nil, want validation error")
	}
	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("VerifyPassword() error = %T, want ValidationError", err)
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	id := uuid.New()
	oldHash, err := bcrypt.GenerateFromPassword([]byte("old-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	st := &userStorageServiceMock{userByID: models.User{ID: id, Password: string(oldHash)}}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})

	err = svc.ChangePassword(context.Background(), id, models.ChangePasswordRequest{OldPassword: "old-pass", NewPassword: "new-pass-123"})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if st.updatedUserReq.Password == nil {
		t.Fatal("updated password is nil")
	}
	if bcrypt.CompareHashAndPassword([]byte(*st.updatedUserReq.Password), []byte("new-pass-123")) != nil {
		t.Fatal("new password hash mismatch")
	}

	err = svc.ChangePassword(context.Background(), id, models.ChangePasswordRequest{OldPassword: "wrong", NewPassword: "new-pass-123"})
	if err == nil {
		t.Fatal("ChangePassword() error = nil, want validation error")
	}
}

func TestUserService_RequestPasswordReset(t *testing.T) {
	id := uuid.New()
	email := "alice@example.com"
	st := &userStorageServiceMock{userByEmail: models.User{ID: id, Email: &email}}
	tk := &userTokenStorageMock{}
	mailer := &userEmailServiceMock{}
	svc := NewUserService(mailer, st, tk, &userAvatarStorageMock{}, &userLoggerMock{})

	if err := svc.RequestPasswordReset(context.Background(), email); err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}
	if tk.saveResetCalls != 1 {
		t.Fatalf("SavePasswordResetToken calls = %d, want 1", tk.saveResetCalls)
	}
	if tk.saveResetUserID != id.String() {
		t.Fatalf("saved reset user ID = %s, want %s", tk.saveResetUserID, id)
	}
	if len(tk.saveResetTokenID) != 64 {
		t.Fatalf("reset token len = %d, want 64", len(tk.saveResetTokenID))
	}
	if mailer.resetCalls != 1 {
		t.Fatalf("SendPasswordResetLetter calls = %d, want 1", mailer.resetCalls)
	}
}

func TestUserService_RequestPasswordReset_SilentCases(t *testing.T) {
	email := "nobody@example.com"
	st := &userStorageServiceMock{userByEmailErr: errors.New("not found")}
	tk := &userTokenStorageMock{}
	mailer := &userEmailServiceMock{}
	svc := NewUserService(mailer, st, tk, &userAvatarStorageMock{}, &userLoggerMock{})
	if err := svc.RequestPasswordReset(context.Background(), email); err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}

	st.userByEmailErr = nil
	st.userByEmail = models.User{ID: uuid.New(), Email: nil}
	if err := svc.RequestPasswordReset(context.Background(), email); err != nil {
		t.Fatalf("RequestPasswordReset() nil-email error = %v", err)
	}
}

func TestUserService_ResetPassword(t *testing.T) {
	userID := uuid.New()
	tk := &userTokenStorageMock{getResetValue: userID.String()}
	st := &userStorageServiceMock{}
	svc := NewUserService(&userEmailServiceMock{}, st, tk, &userAvatarStorageMock{}, &userLoggerMock{})

	if err := svc.ResetPassword(context.Background(), "token", "new-pass-123"); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if st.updatedUserID != userID {
		t.Fatalf("updated user ID = %s, want %s", st.updatedUserID, userID)
	}
	if st.updatedUserReq.Password == nil {
		t.Fatal("updated password is nil")
	}
	if tk.deleteResetCalls != 1 {
		t.Fatalf("DeletePasswordResetToken calls = %d, want 1", tk.deleteResetCalls)
	}
}

func TestUserService_ResetPassword_InvalidToken(t *testing.T) {
	tk := &userTokenStorageMock{getResetErr: errors.New("missing")}
	svc := NewUserService(&userEmailServiceMock{}, &userStorageServiceMock{}, tk, &userAvatarStorageMock{}, &userLoggerMock{})
	err := svc.ResetPassword(context.Background(), "bad", "new-pass-123")
	if err == nil {
		t.Fatal("ResetPassword() error = nil, want validation error")
	}
	var validation svcErr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ResetPassword() error = %T, want ValidationError", err)
	}
}

func TestUserService_Delete(t *testing.T) {
	id := uuid.New()
	st := &userStorageServiceMock{}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})
	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if st.deletedUserID != id {
		t.Fatalf("Delete() user ID = %s, want %s", st.deletedUserID, id)
	}
}

func TestUserService_Register_CreateUserError(t *testing.T) {
	st := &userStorageServiceMock{createErr: errors.New("duplicate")}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, &userAvatarStorageMock{}, &userLoggerMock{})
	_, err := svc.Register(context.Background(), models.RegisterUserRequest{
		Name:     "John",
		Username: "johnny",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("Register() error = nil, want storage error")
	}
}

func TestUserService_VerifyEmail_ParseAndStorageErrors(t *testing.T) {
	svc := NewUserService(&userEmailServiceMock{}, &userStorageServiceMock{}, &userTokenStorageMock{getEmailValue: "not-uuid"}, &userAvatarStorageMock{}, &userLoggerMock{})
	err := svc.VerifyEmail(context.Background(), "token")
	if err == nil {
		t.Fatal("VerifyEmail() error = nil, want parse error")
	}

	userID := uuid.New()
	st := &userStorageServiceMock{setVerifiedErr: errors.New("db failed")}
	svc = NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{getEmailValue: userID.String()}, &userAvatarStorageMock{}, &userLoggerMock{})
	err = svc.VerifyEmail(context.Background(), "token")
	if err == nil {
		t.Fatal("VerifyEmail() error = nil, want storage error")
	}
}

func TestUserService_UpdateAvatar_UploadError(t *testing.T) {
	id := uuid.New()
	st := &userStorageServiceMock{userByID: models.User{ID: id}}
	s3 := &userAvatarStorageMock{uploadErr: errors.New("s3 unavailable")}
	svc := NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{}, s3, &userLoggerMock{})
	err := svc.UpdateAvatar(context.Background(), id, strings.NewReader("x"), 1, "image/png")
	if err == nil {
		t.Fatal("UpdateAvatar() error = nil, want upload error")
	}
}

func TestUserService_RequestPasswordReset_ErrorPaths(t *testing.T) {
	id := uuid.New()
	email := "alice@example.com"
	st := &userStorageServiceMock{userByEmail: models.User{ID: id, Email: &email}}
	tk := &userTokenStorageMock{saveResetErr: errors.New("redis down")}
	svc := NewUserService(&userEmailServiceMock{}, st, tk, &userAvatarStorageMock{}, &userLoggerMock{})
	err := svc.RequestPasswordReset(context.Background(), email)
	if err == nil {
		t.Fatal("RequestPasswordReset() error = nil, want save token error")
	}

	tk = &userTokenStorageMock{}
	mailer := &userEmailServiceMock{resetErr: errors.New("smtp down")}
	svc = NewUserService(mailer, st, tk, &userAvatarStorageMock{}, &userLoggerMock{})
	err = svc.RequestPasswordReset(context.Background(), email)
	if err == nil {
		t.Fatal("RequestPasswordReset() error = nil, want email error")
	}
}

func TestUserService_ResetPassword_ErrorPaths(t *testing.T) {
	svc := NewUserService(&userEmailServiceMock{}, &userStorageServiceMock{}, &userTokenStorageMock{getResetValue: "bad-uuid"}, &userAvatarStorageMock{}, &userLoggerMock{})
	err := svc.ResetPassword(context.Background(), "token", "new-pass")
	if err == nil {
		t.Fatal("ResetPassword() error = nil, want parse error")
	}

	userID := uuid.New()
	st := &userStorageServiceMock{updateErr: errors.New("db failed")}
	svc = NewUserService(&userEmailServiceMock{}, st, &userTokenStorageMock{getResetValue: userID.String()}, &userAvatarStorageMock{}, &userLoggerMock{})
	err = svc.ResetPassword(context.Background(), "token", "new-pass")
	if err == nil {
		t.Fatal("ResetPassword() error = nil, want update error")
	}
}
