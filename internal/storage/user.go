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

type UserStorageImpl struct{ pool *pgxpool.Pool }

func NewUserStorage(pool *pgxpool.Pool) *UserStorageImpl { return &UserStorageImpl{pool: pool} }

func (us *UserStorageImpl) CreateUser(ctx context.Context, user models.User) error {
	if _, err := us.pool.Exec(ctx,
		`INSERT INTO users (id, name, username, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Name, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt,
	); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (us *UserStorageImpl) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User

	if err := us.pool.QueryRow(ctx, `SELECT id, name, username, email, password, created_at, updated_at FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("failed to get user with ID '%s': not found", id)
		}
		return models.User{}, fmt.Errorf("failed to get user with ID '%s': %w", id, err)
	}

	return user, nil
}

func (us *UserStorageImpl) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User

	if err := us.pool.QueryRow(ctx, `SELECT id, name, username, email, password, created_at, updated_at FROM users WHERE username = $1`, username).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("failed to get user with username '%s': not found", username)
		}
		return models.User{}, fmt.Errorf("failed to get user with username '%s': %w", username, err)
	}

	return user, nil
}

func (us *UserStorageImpl) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	if err := us.pool.QueryRow(ctx, `SELECT id, name, username, email, password, created_at, updated_at FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("failed to get user with email '%s': not found", email)
		}
		return models.User{}, fmt.Errorf("failed to get user with email '%s': %w", email, err)
	}

	return user, nil
}

func (us *UserStorageImpl) UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
	var clauses []string
	var args []any
	var index = 1

	if req.Name != nil {
		clauses = append(clauses, fmt.Sprintf("name = $%d", index))
		args = append(args, *req.Name)
		index++
	}
	if req.Username != nil {
		clauses = append(clauses, fmt.Sprintf("username = $%d", index))
		args = append(args, *req.Username)
		index++
	}
	if req.Email != nil {
		clauses = append(clauses, fmt.Sprintf("email = $%d", index))
		args = append(args, *req.Email)
		index++
	}
	if req.Password != nil {
		clauses = append(clauses, fmt.Sprintf("password = $%d", index))
		args = append(args, *req.Password)
		index++
	}

	if len(args) == 0 {
		return nil
	}

	clauses = append(clauses, "updated_at = now()")
	args = append(args, id)

	if result, err := us.pool.Exec(ctx, fmt.Sprintf("UPDATE users SET"+" %s WHERE id = $%d", strings.Join(clauses, ", "), index), args...); err != nil { // "+" to suppress false-positive "<set assignment> expected, got '%'" on "SET %s", "//noinspection ALL" didn't work
		return fmt.Errorf("failed to update user with ID '%s': %w", id, err) //                                                                                                                                                                                   ^
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to update user with ID '%s': not found", id)
	}

	return nil
}

func (us *UserStorageImpl) DeleteUserByID(ctx context.Context, id uuid.UUID) error {
	if result, err := us.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id); err != nil {
		return fmt.Errorf("failed to delete user with ID '%s': %w", id, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete user with ID '%s': not found", id)
	}

	return nil
}
