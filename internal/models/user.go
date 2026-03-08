package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Avatar        *string
	Name          string
	Username      string
	Email         *string
	EmailVerified bool
	Password      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (u User) ToPrivateResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Avatar:        u.Avatar,
		Name:          &u.Name,
		Username:      u.Username,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func (u User) ToPublicResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Avatar:   u.Avatar,
		Name:     &u.Name,
		Username: u.Username,
	}
}

type RegisterUserRequest struct {
	Name     string  `json:"name" binding:"required" example:"Alice"`
	Username string  `json:"username" binding:"required" example:"user421"`
	Email    *string `json:"email" binding:"omitempty,email" example:"alice412@email.com"`
	Password string  `json:"password" binding:"required,min=8" example:"P4s5w0rd"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required" example:"befe235a381306c34ac2e235a565e8d47febd88df9c100fe2cfeab5f7654db75"`
}

type LogInUserRequest struct {
	Username string `json:"username" binding:"required" example:"user421"`
	Password string `json:"password" binding:"required" example:"P4s5w0rd"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDE5Y2QzNDktZDE3Ni03NTYyLWIwM2ItMWRiMjIyM2I5YTAxIiwiaXNzIjoid2lzaGxpc3QuaXRza29zaGtpbi5ydSIsInN1YiI6IjAxOWNkMzQ5LWQxNzYtNzU2Mi1iMDNiLTFkYjIyMjNiOWEwMSIsImF1ZCI6WyJ3aXNobGlzdC5pdHNrb3Noa2luLnJ1Il0sImV4cCI6MTc3MzY3NjgzMywibmJmIjoxNzczMDcyMDMzLCJpYXQiOjE3NzMwNzIwMzMsImp0aSI6IjBiZDU4YWVkLWE1YjYtNDRjNi1iOTdkLWFhYjc5ZTNlZGMxNCJ9._WVHH8QfmA4qbzIyKJKJa6h2uM2jeoRtg8PLJFKSau4"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name" example:"Alice M."`
	Avatar   *string `json:"-"` // Remove?
	Username *string `json:"username" example:"alice421"`
	Email    *string `json:"email" binding:"omitempty,email" example:"alice412+wishlist@email.com"`
	Password *string `json:"-"` // Remove?
}

type ChangePasswordRequest struct {
	OldPassword string `json:"current_password" binding:"required" example:"P4s5w0rd"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"Str0ngerP4s5w0rd"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"alice412@email.com"`
}

type SetNewPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"3b5b0860ed1be5c0fe6b18db6615bd05046b09677aa514a6e46c232cbff1bf7a"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"Str0ngerP4s5w0rd"`
}

type DeleteAccountRequest struct {
	CurrentPassword string `json:"password" binding:"required" example:"P4s5w0rd"`
}

type UserResponse struct {
	ID            uuid.UUID `json:"id" example:"019cd349-d176-7562-b03b-1db2223b9a01"`
	Avatar        *string   `json:"avatar" example:"null"`
	Name          *string   `json:"name" example:"Alice"`
	Username      string    `json:"username" example:"alice421"`
	Email         *string   `json:"email" example:"alice412@email.com"`
	EmailVerified bool      `json:"email_verified" example:"true"`
	CreatedAt     time.Time `json:"created_at" example:"2026-03-08T18:00:00.000000+03:00"`
	UpdatedAt     time.Time `json:"updated_at" example:"2026-03-08T18:03:00.000000+03:00"`
}
