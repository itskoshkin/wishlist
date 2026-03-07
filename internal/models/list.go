package models

import (
	"time"

	"github.com/google/uuid"
)

type List struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Image      *string
	Title      string
	Notes      *string
	IsPublic   bool
	ShareToken string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (l List) ToOwnerResponse() ListResponse {
	return ListResponse{
		ID:         l.ID,
		UserID:     l.UserID,
		Image:      l.Image,
		Title:      l.Title,
		Notes:      l.Notes,
		IsPublic:   l.IsPublic,
		ShareToken: l.ShareToken,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}

func (l List) ToViewerResponse(currentUserID *uuid.UUID) ListResponse {
	return ListResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		Image:     l.Image,
		Title:     l.Title,
		Notes:     l.Notes,
		IsPublic:  l.IsPublic,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

type CreateListRequest struct {
	Title string  `json:"title" binding:"required"`
	Notes *string `json:"notes"`
}

type UpdateListRequest struct {
	Image    *string `json:"image"`
	Title    *string `json:"title"`
	Notes    *string `json:"notes"`
	IsPublic *bool   `json:"is_public"`
}

type ListResponse struct {
	ID         uuid.UUID      `json:"id"`
	UserID     uuid.UUID      `json:"user_id"`
	Image      *string        `json:"image"`
	Title      string         `json:"title"`
	Notes      *string        `json:"notes,omitempty"`
	IsPublic   bool           `json:"is_public"`
	ShareToken string         `json:"share_token"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Wishes     []WishResponse `json:"wishes,omitempty"`
}
