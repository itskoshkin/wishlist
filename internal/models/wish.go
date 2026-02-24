package models

import (
	"time"

	"github.com/google/uuid"
)

type Wish struct {
	ID     uuid.UUID `json:"id"`
	ListID uuid.UUID `json:"list_id"`
	//Image *string `json:"image"`
	Title string  `json:"title"`
	Notes *string `json:"notes"`
	Link  *string `json:"link"`
	//Price     *int64    `json:"price,omitempty"`
	//Reserved  *uuid.UUID      `json:"reserved"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWishRequest struct {
	ListID uuid.UUID `json:"list_id" binding:"required"`
	Title  string    `json:"title" binding:"required"`
	Notes  *string   `json:"notes"`
	Link   *string   `json:"link" binding:"omitempty,url"`
}

type UpdateWishRequest struct {
	Title *string `json:"title"`
	Notes *string `json:"notes"`
	Link  *string `json:"link" binding:"omitempty,url"`
}

type WishResponse struct {
	ID        uuid.UUID `json:"id"`
	ListID    uuid.UUID `json:"list_id"`
	Title     string    `json:"title"`
	Notes     *string   `json:"notes,omitempty"`
	Link      *string   `json:"link,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
