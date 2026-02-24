package models

import (
	"time"

	"github.com/google/uuid"
)

type List struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	//Image *string `json:"image"`
	Title     string    `json:"title"`
	Notes     *string   `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateListRequest struct {
	Title string  `json:"title" binding:"required"`
	Notes *string `json:"notes"`
}

type UpdateListRequest struct {
	Title *string `json:"title"`
	Notes *string `json:"notes"`
}

type ListResponse struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	Title     string         `json:"title"`
	Notes     *string        `json:"notes,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Wishes    []WishResponse `json:"wishes,omitempty"`
}
