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
)

type ListStorageImpl struct{ pool *pgxpool.Pool }

func NewListStorage(pool *pgxpool.Pool) *ListStorageImpl { return &ListStorageImpl{pool: pool} }

func (s *ListStorageImpl) CreateList(ctx context.Context, list models.List) error {
	if _, err := s.pool.Exec(ctx,
		`INSERT INTO lists (id, user_id, image, title, notes, is_public, share_token, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		list.ID, list.UserID, list.Image, list.Title, list.Notes, list.IsPublic, list.ShareToken, list.CreatedAt, list.UpdatedAt,
	); err != nil {
		return fmt.Errorf("failed to create list: %w", err)
	}

	return nil
}

func (s *ListStorageImpl) GetListByID(ctx context.Context, id uuid.UUID) (models.List, error) {
	var list models.List

	if err := s.pool.QueryRow(ctx, `SELECT id, user_id, image, title, notes, is_public, share_token, created_at, updated_at FROM lists WHERE id = $1`, id).Scan(
		&list.ID, &list.UserID, &list.Image, &list.Title, &list.Notes, &list.IsPublic, &list.ShareToken, &list.CreatedAt, &list.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.List{}, fmt.Errorf("failed to get list with ID '%s': not found", id)
		}
		return models.List{}, fmt.Errorf("failed to get list with ID '%s': %w", id, err)
	}

	return list, nil
}

func (s *ListStorageImpl) GetListBySharedLink(ctx context.Context, slug string) (models.List, error) {
	var list models.List

	if err := s.pool.QueryRow(ctx, `SELECT id, user_id, image, title, notes, is_public, share_token, created_at, updated_at FROM lists WHERE share_token = $1`, slug).Scan(
		&list.ID, &list.UserID, &list.Image, &list.Title, &list.Notes, &list.IsPublic, &list.ShareToken, &list.CreatedAt, &list.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.List{}, fmt.Errorf("failed to get list with slug '%s': not found", slug)
		}
		return models.List{}, fmt.Errorf("failed to get list with slug '%s': %w", slug, err)
	}

	return list, nil
}

func (s *ListStorageImpl) GetListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, image, title, notes, is_public, share_token, created_at, updated_at FROM lists WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all lists for user with ID '%s': %w", userID, err)
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var list models.List
		if err = rows.Scan(&list.ID, &list.UserID, &list.Image, &list.Title, &list.Notes, &list.IsPublic, &list.ShareToken, &list.CreatedAt, &list.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan list: %w", err)
		}
		lists = append(lists, list)
	}

	return lists, rows.Err()
}

func (s *ListStorageImpl) GetPublicListsByUserID(ctx context.Context, userID uuid.UUID) ([]models.List, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, image, title, notes, is_public, share_token, created_at, updated_at FROM lists WHERE user_id = $1 AND is_public = true ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public lists for user with ID '%s': %w", userID, err)
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var list models.List
		if err = rows.Scan(&list.ID, &list.UserID, &list.Image, &list.Title, &list.Notes, &list.IsPublic, &list.ShareToken, &list.CreatedAt, &list.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan list: %w", err)
		}
		lists = append(lists, list)
	}

	return lists, rows.Err()
}

func (s *ListStorageImpl) UpdateListByID(ctx context.Context, listID uuid.UUID, req models.UpdateListRequest) error {
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
	if req.IsPublic != nil {
		clauses = append(clauses, fmt.Sprintf("is_public = $%d", index))
		args = append(args, *req.IsPublic)
		index++
	}
	if len(args) == 0 {
		return nil
	}

	clauses = append(clauses, "updated_at = now()")
	args = append(args, listID)

	if result, err := s.pool.Exec(ctx, fmt.Sprintf("UPDATE lists SET"+" %s WHERE id = $%d", strings.Join(clauses, ", "), index), args...); err != nil {
		return fmt.Errorf("failed to update list with ID '%s': %w", listID, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to update list with ID '%s': not found", listID)
	}

	return nil
}

func (s *ListStorageImpl) RotateSharedLink(ctx context.Context, listID uuid.UUID, token string) error {
	if result, err := s.pool.Exec(ctx, "UPDATE lists SET share_token = $1, updated_at = now() WHERE id = $2", token, listID); err != nil {
		return fmt.Errorf("failed to update share token for list with ID '%s': %w", listID, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to update share token for list with ID '%s': not found", listID)
	}

	return nil
}

func (s *ListStorageImpl) DeleteListByID(ctx context.Context, listID uuid.UUID) error {
	if result, err := s.pool.Exec(ctx, "DELETE FROM lists WHERE id = $1", listID); err != nil {
		return fmt.Errorf("failed to delete list with ID '%s': %w", listID, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete list with ID '%s': not found", listID)
	}

	return nil
}
