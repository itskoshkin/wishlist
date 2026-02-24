package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID
	//Avatar *string //TODO: Add support for user avatars (S3)
	Name      string
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) ToPrivateResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      &u.Name,
		Username:  u.Username,
		Email:     &u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (u User) ToPublicResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Name:     &u.Name,
		Username: u.Username,
	}
}

type RegisterUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LogInUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Username *string `json:"username"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"-"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"current_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type DeleteAccountRequest struct {
	CurrentPassword string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID uuid.UUID `json:"id"`
	//Avatar *string `json:"image"` //TODO: Add support for user avatars (S3)
	Name      *string   `json:"name"`
	Username  string    `json:"username"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
