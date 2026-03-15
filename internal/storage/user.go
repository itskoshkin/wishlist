package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"wishlist/internal/models"
	"wishlist/internal/services/errors"
)

type UserStorageImpl struct{ pool *pgxpool.Pool }

func NewUserStorage(pool *pgxpool.Pool) *UserStorageImpl { return &UserStorageImpl{pool: pool} }

func (us *UserStorageImpl) CreateUser(ctx context.Context, user models.User) error {
	if _, err := us.pool.Exec(ctx,
		`INSERT INTO users (id, name, username, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Name, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt,
	); err != nil {
		if mappedErr := mapUserWriteError(err); mappedErr != err {
			return mappedErr
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (us *UserStorageImpl) SetUserEmailAsVerified(ctx context.Context, id uuid.UUID) error {
	result, err := us.pool.Exec(ctx,
		"UPDATE users SET email_verified = true, updated_at = now() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to verify email for user '%s': %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to verify email for user '%s': not found", id)
	}

	return nil
}

func (us *UserStorageImpl) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User

	if err := us.pool.QueryRow(ctx, `SELECT id, avatar, name, username, email, email_verified, password, created_at, updated_at FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Avatar, &user.Name, &user.Username, &user.Email, &user.EmailVerified, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, svcErr.NotFoundError{Entity: "user", Field: "id", Value: id.String()}
		}
		return models.User{}, fmt.Errorf("failed to get user with ID '%s': %w", id, err)
	}

	return user, nil
}

func (us *UserStorageImpl) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User

	if err := us.pool.QueryRow(ctx, `SELECT id, avatar, name, username, email, email_verified, password, created_at, updated_at FROM users WHERE lower(username) = lower($1)`, username).Scan(
		&user.ID, &user.Avatar, &user.Name, &user.Username, &user.Email, &user.EmailVerified, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, svcErr.NotFoundError{Entity: "user", Field: "username", Value: username}
		}
		return models.User{}, fmt.Errorf("failed to get user with username '%s': %w", username, err)
	}

	return user, nil
}

func (us *UserStorageImpl) SearchUsersByUsername(ctx context.Context, query string, limit int) ([]models.User, error) {
	search := strings.TrimSpace(query)
	if search == "" {
		return []models.User{}, nil
	}
	if limit <= 0 {
		limit = 5
	}

	//noinspection SqlRedundantOrderingDirection
	rows, err := us.pool.Query(ctx, `SELECT id, avatar, name, username, email, email_verified, password, created_at, updated_at FROM users WHERE lower(username) LIKE $1 ORDER BY CASE WHEN lower(username) LIKE $2 THEN 0 ELSE 1 END, lower(username) ASC LIMIT $3`, "%"+search+"%", search+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users with query '%s': %w", search, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err = rows.Scan(
			&user.ID, &user.Avatar, &user.Name, &user.Username, &user.Email, &user.EmailVerified, &user.Password, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan searched user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (us *UserStorageImpl) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	if err := us.pool.QueryRow(ctx, `SELECT id, avatar, name, username, email, email_verified, password, created_at, updated_at FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Avatar, &user.Name, &user.Username, &user.Email, &user.EmailVerified, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, svcErr.NotFoundError{Entity: "user", Field: "email", Value: email}
		}
		return models.User{}, fmt.Errorf("failed to get user with email '%s': %w", email, err)
	}

	return user, nil
}

func (us *UserStorageImpl) UpdateUserByID(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) error {
	var clauses []string
	var args []any
	var index = 1

	if req.Avatar != nil {
		clauses = append(clauses, fmt.Sprintf("avatar = $%d", index))
		args = append(args, *req.Avatar)
		index++
	}
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
		clauses = append(clauses, "email_verified = false")
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
		if mappedErr := mapUserWriteError(err); mappedErr != err {
			return mappedErr
		}
		return fmt.Errorf("failed to update user with ID '%s': %w", id, err) //                                                                                                                                                                                   ^
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to update user with ID '%s': not found", id)
	}

	return nil
}

func (us *UserStorageImpl) RemoveUserAvatar(ctx context.Context, id uuid.UUID) error {
	if result, err := us.pool.Exec(ctx, "UPDATE users SET avatar = NULL, updated_at = now() WHERE id = $1", id); err != nil {
		return fmt.Errorf("failed to remove avatar for user with ID '%s': %w", id, err)
	} else if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to remove avatar for user with ID '%s': not found", id)
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

func mapUserWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "users_username_key", "users_username_lower_key":
			return svcErr.ConflictError{Message: "username is already taken"}
		case "users_email_key":
			return svcErr.ConflictError{Message: "email is already in use"}
		}
		return svcErr.ConflictError{Message: "user with these credentials already exists"}
	}
	return err
}
