package models

import (
	"time"

	"github.com/google/uuid"
)

type Wish struct {
	ID         uuid.UUID
	ListID     uuid.UUID
	Image      *string
	Title      string
	Notes      *string
	Link       *string
	Price      *int64
	Currency   *string
	ReservedBy *uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (w Wish) ToOwnerResponse() WishResponse {
	return WishResponse{
		ID:         w.ID,
		ListID:     w.ListID,
		Image:      w.Image,
		Title:      w.Title,
		Notes:      w.Notes,
		Link:       w.Link,
		Price:      w.Price,
		Currency:   w.Currency,
		Reserved:   w.ReservedBy != nil,
		ReservedBy: nil, // Surprise
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

func (w Wish) ToViewerResponse(requestedByUserID *uuid.UUID) WishResponse {
	var reservedBy *uuid.UUID
	if w.ReservedBy != nil && requestedByUserID != nil && *w.ReservedBy == *requestedByUserID {
		reservedBy = w.ReservedBy
	}

	return WishResponse{
		ID:         w.ID,
		ListID:     w.ListID,
		Image:      w.Image,
		Title:      w.Title,
		Notes:      w.Notes,
		Link:       w.Link,
		Price:      w.Price,
		Currency:   w.Currency,
		Reserved:   w.ReservedBy != nil,
		ReservedBy: reservedBy, // Only show if user who requested == user who reserved
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

type CreateWishRequest struct {
	Image    *string `json:"image,omitempty"`
	Title    string  `json:"title" binding:"required"`
	Notes    *string `json:"notes"`
	Link     *string `json:"link" binding:"omitempty,url"`
	Price    *int64  `json:"price"`
	Currency *string `json:"currency"`
}

type UpdateWishRequest struct {
	Image    *string `json:"image,omitempty"`
	Title    *string `json:"title"`
	Notes    *string `json:"notes"`
	Link     *string `json:"link" binding:"omitempty,url"`
	Price    *int64  `json:"price"`
	Currency *string `json:"currency"`
}

type WishResponse struct {
	ID         uuid.UUID  `json:"id"`
	ListID     uuid.UUID  `json:"list_id"`
	Image      *string    `json:"image,omitempty"`
	Title      string     `json:"title"`
	Notes      *string    `json:"notes,omitempty"`
	Link       *string    `json:"link,omitempty"`
	Price      *int64     `json:"price,omitempty"`
	Currency   *string    `json:"currency,omitempty"`
	Reserved   bool       `json:"reserved"`
	ReservedBy *uuid.UUID `json:"reserved_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
