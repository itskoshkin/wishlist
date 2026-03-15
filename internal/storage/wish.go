package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
)

type WishStorageImpl struct{ pool *pgxpool.Pool }

func NewWishStorage(pool *pgxpool.Pool) *WishStorageImpl { return &WishStorageImpl{pool: pool} }

func (s *WishStorageImpl) CreateWish(ctx context.Context, wish models.Wish) error {
	if _, err := s.pool.Exec(ctx, `INSERT INTO wishes (id, list_id, image, title, notes, link, price, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		wish.ID, wish.ListID, wish.Image, wish.Title, wish.Notes, wish.Link, wish.Price, wish.Currency, wish.CreatedAt, wish.UpdatedAt,
	); err != nil {
		return fmt.Errorf("failed to create wish: %w", err)
	}

	return nil
}

func (s *WishStorageImpl) GetWishByID(ctx context.Context, wishID uuid.UUID) (models.Wish, error) {
	var wish models.Wish

	if err := s.pool.QueryRow(ctx, `SELECT id, list_id, image, title, notes, link, price, currency, reserved_by, created_at, updated_at FROM wishes WHERE id = $1`, wishID).Scan(
		&wish.ID, &wish.ListID, &wish.Image, &wish.Title, &wish.Notes, &wish.Link, &wish.Price, &wish.Currency, &wish.ReservedBy, &wish.CreatedAt, &wish.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Wish{}, svcErr.NotFoundError{Entity: "wish", Field: "id", Value: wishID.String()}
		}
		return models.Wish{}, fmt.Errorf("failed to get wish with ID '%s': %w", wishID, err)
	}

	return wish, nil
}

func (s *WishStorageImpl) GetWishesByListID(ctx context.Context, listID uuid.UUID) ([]models.Wish, error) {
	//noinspection SqlRedundantOrderingDirection
	rows, err := s.pool.Query(ctx, `SELECT id, list_id, image, title, notes, link, price, currency, reserved_by, created_at, updated_at FROM wishes WHERE list_id = $1 ORDER BY created_at ASC`, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishes for list with ID '%s': %w", listID, err)
	}
	defer rows.Close()

	var wishes []models.Wish
	for rows.Next() {
		var wish models.Wish
		if err = rows.Scan(&wish.ID, &wish.ListID, &wish.Image, &wish.Title, &wish.Notes, &wish.Link, &wish.Price, &wish.Currency, &wish.ReservedBy, &wish.CreatedAt, &wish.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan wish: %w", err)
		}
		wishes = append(wishes, wish)
	}

	return wishes, rows.Err()
}

// noinspection DuplicatedCode
func (s *WishStorageImpl) UpdateWishByID(ctx context.Context, wishID uuid.UUID, req models.UpdateWishRequest) error {
	var clauses []string
	var args []any
	var index = 1

	if req.Image != nil {
		clauses = append(clauses, fmt.Sprintf("image = $%d", index))
		args = append(args, *req.Image)
		index++
	}
	if req.Title != nil {
		clauses = append(clauses, fmt.Sprintf("title = $%d", index))
		args = append(args, *req.Title)
		index++
	}
	if req.Notes != nil {
		clauses = append(clauses, fmt.Sprintf("notes = $%d", index))
		args = append(args, *req.Notes)
		index++
	}
	if req.Link != nil {
		clauses = append(clauses, fmt.Sprintf("link = $%d", index))
		args = append(args, *req.Link)
		index++
	}
	if req.Price != nil {
		clauses = append(clauses, fmt.Sprintf("price = $%d", index))
		args = append(args, *req.Price)
		index++
	}
	if req.Currency != nil {
		clauses = append(clauses, fmt.Sprintf("currency = $%d", index))
		args = append(args, *req.Currency)
		index++
	}
	if len(args) == 0 {
		return nil
	}

	clauses = append(clauses, "updated_at = now()")
	args = append(args, wishID)

	if result, err := s.pool.Exec(ctx, fmt.Sprintf("UPDATE wishes SET"+" %s WHERE id = $%d", strings.Join(clauses, ", "), index), args...); err != nil {
		return fmt.Errorf("failed to update wish with ID '%s': %w", wishID, err)
	} else if result.RowsAffected() == 0 {
		return svcErr.NotFoundError{Entity: "wish", Field: "id", Value: wishID.String()}
	}

	return nil
}

func (s *WishStorageImpl) ReserveWish(ctx context.Context, wishID, userID uuid.UUID) error {
	if result, err := s.pool.Exec(ctx, `UPDATE wishes SET reserved_by = $1, updated_at = now() WHERE id = $2 AND reserved_by IS NULL`, userID, wishID); err != nil {
		return fmt.Errorf("failed to reserve wish with ID '%s': %w", wishID, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to reserve wish with ID '%s': already reserved or not found", wishID)
	}

	return nil
}

func (s *WishStorageImpl) ReleaseWish(ctx context.Context, wishID, userID uuid.UUID) error {
	if result, err := s.pool.Exec(ctx, `UPDATE wishes SET reserved_by = NULL, updated_at = now() WHERE id = $1 AND reserved_by = $2`, wishID, userID); err != nil {
		return fmt.Errorf("failed to release wish with ID '%s': %w", wishID, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to release wish with ID '%s': not reserved by you or not found", wishID)
	}

	return nil
}

func (s *WishStorageImpl) DeleteWishByID(ctx context.Context, wishID uuid.UUID) error {
	if result, err := s.pool.Exec(ctx, "DELETE FROM wishes WHERE id = $1", wishID); err != nil {
		return fmt.Errorf("failed to delete wish with ID '%s': %w", wishID, err)
	} else if result.RowsAffected() == 0 {
		return svcErr.NotFoundError{Entity: "wish", Field: "id", Value: wishID.String()}
	}

	return nil
}
