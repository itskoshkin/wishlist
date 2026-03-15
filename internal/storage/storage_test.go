//go:build integration

package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	minio "github.com/minio/minio-go/v7"
	redisgo "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"wishlist/internal/config"
	"wishlist/internal/models"
	minioPkg "wishlist/pkg/minio"
	"wishlist/pkg/postgres"
	redisPkg "wishlist/pkg/redis"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := postgres.Config{
		Host:     getEnv("TEST_DB_HOST", getEnv("DB_HOST", "localhost")),
		Port:     getEnv("TEST_DB_PORT", getEnv("DB_PORT", "5432")),
		User:     getEnv("TEST_DB_USER", getEnv("DB_USER", "postgres")),
		Password: getEnv("TEST_DB_PASSWORD", getEnv("DB_PASSWORD", "4221")),
		Database: getEnv("TEST_DB_NAME", getEnv("DB_NAME", "wishlist")),
		SSLMode:  getEnv("TEST_DB_SSLMODE", getEnv("DB_SSLMODE", "disable")),
	}

	pool, err := postgres.NewInstance(ctx, cfg)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}

	ensureSchema(t, pool)
	resetDB(t, pool)
	return pool
}

func ensureSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			avatar TEXT,
			name TEXT NOT NULL,
			username TEXT NOT NULL,
			email TEXT UNIQUE,
			email_verified BOOLEAN NOT NULL DEFAULT FALSE,
			password TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_key ON users ((lower(username)));`,
		`CREATE TABLE IF NOT EXISTS lists (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			image TEXT,
			title TEXT NOT NULL,
			notes TEXT,
			is_public BOOLEAN NOT NULL DEFAULT TRUE,
			slug VARCHAR(32) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT lists_slug_len CHECK (char_length(slug) = 32)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_lists_user_id_created_at_desc ON lists (user_id, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS wishes (
			id UUID PRIMARY KEY,
			list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
			image TEXT,
			title TEXT NOT NULL,
			notes TEXT,
			link TEXT,
			price BIGINT,
			currency VARCHAR(8),
			reserved_by UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT wishes_price_non_negative CHECK (price IS NULL OR price >= 0)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_wishes_list_id_created_at_asc ON wishes (list_id, created_at ASC);`,
		`CREATE INDEX IF NOT EXISTS idx_wishes_reserved_by ON wishes (reserved_by);`,
	}

	for _, stmt := range stmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("ensureSchema failed: %v", err)
		}
	}
}

func resetDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, "TRUNCATE TABLE wishes, lists, users CASCADE"); err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
}

func mustRedis(t *testing.T) *redisgo.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbNum := 0
	if v := getEnv("TEST_REDIS_DB", getEnv("REDIS_DB", "0")); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			dbNum = i
		}
	}

	cfg := redisPkg.Config{
		Addr:     getEnv("TEST_REDIS_HOST", getEnv("REDIS_HOST", "localhost")),
		Port:     getEnv("TEST_REDIS_PORT", getEnv("REDIS_PORT", "6379")),
		Password: getEnv("TEST_REDIS_PASSWORD", getEnv("REDIS_PASSWORD", "4221")),
		Database: dbNum,
	}

	client, err := redisPkg.NewClient(ctx, cfg)
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	return client
}

func mustMinio(t *testing.T) (*minio.Client, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bucket := getEnv("TEST_MINIO_BUCKET", getEnv("MINIO_BUCKET", "wishlist"))
	cfg := minioPkg.Config{
		Endpoint:        getEnv("TEST_MINIO_ENDPOINT", getEnv("MINIO_ENDPOINT", "localhost:9000")),
		AccessKeyID:     getEnv("TEST_MINIO_ACCESS_KEY", getEnv("MINIO_ACCESS_KEY", "minioadmin")),
		SecretAccessKey: getEnv("TEST_MINIO_SECRET_KEY", getEnv("MINIO_SECRET_KEY", "minioadmin")),
		UseSSL:          getEnv("TEST_MINIO_USE_SSL", getEnv("MINIO_USE_SSL", "false")) == "true",
		BucketName:      bucket,
	}

	client, err := minioPkg.NewClient(ctx, cfg)
	if err != nil {
		t.Skipf("minio unavailable: %v", err)
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		t.Skipf("minio bucket check failed: %v", err)
	}
	if !exists {
		if err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			t.Skipf("minio create bucket failed: %v", err)
		}
	}

	return client, bucket
}

func TestUserStorage_Integration(t *testing.T) {
	pool := mustPostgres(t)
	resetDB(t, pool)
	store := NewUserStorage(pool)

	ctx := context.Background()
	userID := uuid.New()
	email := "user@example.com"
	user := models.User{
		ID:        userID,
		Name:      "Alice",
		Username:  "alice",
		Email:     &email,
		Password:  "hash",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := store.GetUserByID(ctx, userID)
	if err != nil || got.Username != "alice" {
		t.Fatalf("GetUserByID() error=%v user=%+v", err, got)
	}

	got, err = store.GetUserByUsername(ctx, "alice")
	if err != nil || got.ID != userID {
		t.Fatalf("GetUserByUsername() error=%v user=%+v", err, got)
	}

	got, err = store.GetUserByEmail(ctx, email)
	if err != nil || got.ID != userID {
		t.Fatalf("GetUserByEmail() error=%v user=%+v", err, got)
	}

	newName := "Alice Updated"
	if err = store.UpdateUserByID(ctx, userID, models.UpdateUserRequest{Name: &newName}); err != nil {
		t.Fatalf("UpdateUserByID() error = %v", err)
	}
	got, _ = store.GetUserByID(ctx, userID)
	if got.Name != newName {
		t.Fatalf("UpdateUserByID() name = %s", got.Name)
	}

	if err = store.SetUserEmailAsVerified(ctx, userID); err != nil {
		t.Fatalf("SetUserEmailAsVerified() error = %v", err)
	}
	got, _ = store.GetUserByID(ctx, userID)
	if !got.EmailVerified {
		t.Fatal("email_verified = false")
	}

	avatar := "http://minio/bucket/avatars/user/file"
	if err = store.UpdateUserByID(ctx, userID, models.UpdateUserRequest{Avatar: &avatar}); err != nil {
		t.Fatalf("UpdateUserByID(avatar) error = %v", err)
	}
	if err = store.RemoveUserAvatar(ctx, userID); err != nil {
		t.Fatalf("RemoveUserAvatar() error = %v", err)
	}
	got, _ = store.GetUserByID(ctx, userID)
	if got.Avatar != nil {
		t.Fatal("avatar not cleared")
	}

	if err = store.DeleteUserByID(ctx, userID); err != nil {
		t.Fatalf("DeleteUserByID() error = %v", err)
	}
	if _, err = store.GetUserByID(ctx, userID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestListStorage_Integration(t *testing.T) {
	pool := mustPostgres(t)
	resetDB(t, pool)
	users := NewUserStorage(pool)
	lists := NewListStorage(pool)

	ctx := context.Background()
	userID := uuid.New()
	user := models.User{ID: userID, Name: "Bob", Username: "bob", Password: "hash", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := users.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	list := models.List{
		ID:         uuid.New(),
		UserID:     userID,
		Title:      "List",
		Notes:      nil,
		IsPublic:   true,
		Slug: "12345678901234567890123456789012",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := lists.CreateList(ctx, list); err != nil {
		t.Fatalf("CreateList() error = %v", err)
	}

	got, err := lists.GetListByID(ctx, list.ID)
	if err != nil || got.Title != "List" {
		t.Fatalf("GetListByID() error=%v list=%+v", err, got)
	}

	got, err = lists.GetListBySharedLink(ctx, list.Slug)
	if err != nil || got.ID != list.ID {
		t.Fatalf("GetListBySharedLink() error=%v list=%+v", err, got)
	}

	all, err := lists.GetListsByUserID(ctx, userID)
	if err != nil || len(all) != 1 {
		t.Fatalf("GetListsByUserID() error=%v len=%d", err, len(all))
	}

	public, err := lists.GetPublicListsByUserID(ctx, userID)
	if err != nil || len(public) != 1 {
		t.Fatalf("GetPublicListsByUserID() error=%v len=%d", err, len(public))
	}

	newTitle := "Updated"
	if err = lists.UpdateListByID(ctx, list.ID, models.UpdateListRequest{Title: &newTitle}); err != nil {
		t.Fatalf("UpdateListByID() error = %v", err)
	}
	got, _ = lists.GetListByID(ctx, list.ID)
	if got.Title != newTitle {
		t.Fatalf("UpdateListByID() title = %s", got.Title)
	}

	if err = lists.RotateSharedLink(ctx, list.ID, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); err != nil {
		t.Fatalf("RotateSharedLink() error = %v", err)
	}

	if err = lists.DeleteListByID(ctx, list.ID); err != nil {
		t.Fatalf("DeleteListByID() error = %v", err)
	}
	if _, err = lists.GetListByID(ctx, list.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestWishStorage_Integration(t *testing.T) {
	pool := mustPostgres(t)
	resetDB(t, pool)
	users := NewUserStorage(pool)
	lists := NewListStorage(pool)
	wishes := NewWishStorage(pool)

	ctx := context.Background()
	ownerID := uuid.New()
	otherID := uuid.New()

	if err := users.CreateUser(ctx, models.User{ID: ownerID, Name: "Owner", Username: "owner", Password: "hash", CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("CreateUser(owner) error = %v", err)
	}
	if err := users.CreateUser(ctx, models.User{ID: otherID, Name: "Other", Username: "other", Password: "hash", CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("CreateUser(other) error = %v", err)
	}

	list := models.List{
		ID:         uuid.New(),
		UserID:     ownerID,
		Title:      "List",
		IsPublic:   true,
		Slug: "12345678901234567890123456789012",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := lists.CreateList(ctx, list); err != nil {
		t.Fatalf("CreateList() error = %v", err)
	}

	price := int64(100)
	currency := "USD"
	wish := models.Wish{
		ID:        uuid.New(),
		ListID:    list.ID,
		Title:     "Wish",
		Price:     &price,
		Currency:  &currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := wishes.CreateWish(ctx, wish); err != nil {
		t.Fatalf("CreateWish() error = %v", err)
	}

	got, err := wishes.GetWishByID(ctx, wish.ID)
	if err != nil || got.Title != "Wish" {
		t.Fatalf("GetWishByID() error=%v wish=%+v", err, got)
	}

	listWishes, err := wishes.GetWishesByListID(ctx, list.ID)
	if err != nil || len(listWishes) != 1 {
		t.Fatalf("GetWishesByListID() error=%v len=%d", err, len(listWishes))
	}

	newTitle := "Updated"
	if err = wishes.UpdateWishByID(ctx, wish.ID, models.UpdateWishRequest{Title: &newTitle}); err != nil {
		t.Fatalf("UpdateWishByID() error = %v", err)
	}

	if err = wishes.ReserveWish(ctx, wish.ID, otherID); err != nil {
		t.Fatalf("ReserveWish() error = %v", err)
	}
	if err = wishes.ReserveWish(ctx, wish.ID, ownerID); err == nil {
		t.Fatal("expected reserve error on already reserved")
	}

	if err = wishes.ReleaseWish(ctx, wish.ID, ownerID); err == nil {
		t.Fatal("expected release error for wrong user")
	}
	if err = wishes.ReleaseWish(ctx, wish.ID, otherID); err != nil {
		t.Fatalf("ReleaseWish() error = %v", err)
	}

	if err = wishes.DeleteWishByID(ctx, wish.ID); err != nil {
		t.Fatalf("DeleteWishByID() error = %v", err)
	}
	if _, err = wishes.GetWishByID(ctx, wish.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestWishStorage_InvalidListID_Integration(t *testing.T) {
	pool := mustPostgres(t)
	resetDB(t, pool)
	wishes := NewWishStorage(pool)

	ctx := context.Background()
	badListID := uuid.New()
	price := int64(10)
	wish := models.Wish{
		ID:        uuid.New(),
		ListID:    badListID,
		Title:     "Invalid",
		Price:     &price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := wishes.CreateWish(ctx, wish); err == nil {
		t.Fatal("expected error on CreateWish with invalid list ID")
	}
}

func TestCascadeDelete_Integration(t *testing.T) {
	pool := mustPostgres(t)
	resetDB(t, pool)
	users := NewUserStorage(pool)
	lists := NewListStorage(pool)
	wishes := NewWishStorage(pool)

	ctx := context.Background()
	userID := uuid.New()
	if err := users.CreateUser(ctx, models.User{ID: userID, Name: "Cascade", Username: "cascade", Password: "hash", CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	list := models.List{
		ID:         uuid.New(),
		UserID:     userID,
		Title:      "Cascade list",
		IsPublic:   true,
		Slug: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := lists.CreateList(ctx, list); err != nil {
		t.Fatalf("CreateList() error = %v", err)
	}

	wish := models.Wish{
		ID:        uuid.New(),
		ListID:    list.ID,
		Title:     "Cascade wish",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := wishes.CreateWish(ctx, wish); err != nil {
		t.Fatalf("CreateWish() error = %v", err)
	}

	if err := users.DeleteUserByID(ctx, userID); err != nil {
		t.Fatalf("DeleteUserByID() error = %v", err)
	}
	if _, err := lists.GetListByID(ctx, list.ID); err == nil {
		t.Fatal("expected list deleted on user delete")
	}
	if _, err := wishes.GetWishByID(ctx, wish.ID); err == nil {
		t.Fatal("expected wish deleted on user delete")
	}
}

func TestTokenStorage_RedisIntegration(t *testing.T) {
	client := mustRedis(t)
	defer func() { _ = client.Close() }()

	viper.Reset()
	viper.Set(config.PwdResetTokenTTL, "1h")
	viper.Set(config.EmailVerifyTokenTTL, "1h")

	ts := NewTokenStorage(client)
	ctx := context.Background()

	if err := ts.SaveEmailVerificationToken(ctx, "token1", "user1"); err != nil {
		t.Fatalf("SaveEmailVerificationToken() error = %v", err)
	}
	if val, err := ts.GetEmailVerificationToken(ctx, "token1"); err != nil || val != "user1" {
		t.Fatalf("GetEmailVerificationToken() error=%v val=%s", err, val)
	}
	if err := ts.DeleteEmailVerificationToken(ctx, "token1"); err != nil {
		t.Fatalf("DeleteEmailVerificationToken() error = %v", err)
	}

	revoked, err := ts.CheckIfAuthTokenRevoked(ctx, "tokenX")
	if err != nil || revoked {
		t.Fatalf("CheckIfAuthTokenRevoked() err=%v revoked=%v", err, revoked)
	}
	if err = ts.RevokeAuthTokens(ctx, "tokenX", time.Minute); err != nil {
		t.Fatalf("RevokeAuthTokens() error = %v", err)
	}
	if revoked, err = ts.CheckIfAuthTokenRevoked(ctx, "tokenX"); err != nil || !revoked {
		t.Fatalf("CheckIfAuthTokenRevoked() err=%v revoked=%v", err, revoked)
	}

	if err := ts.SavePasswordResetToken(ctx, "token2", "user2"); err != nil {
		t.Fatalf("SavePasswordResetToken() error = %v", err)
	}
	if val, err := ts.GetPasswordResetToken(ctx, "token2"); err != nil || val != "user2" {
		t.Fatalf("GetPasswordResetToken() error=%v val=%s", err, val)
	}
	if err := ts.DeletePasswordResetToken(ctx, "token2"); err != nil {
		t.Fatalf("DeletePasswordResetToken() error = %v", err)
	}
}

func TestMinioStorage_Integration(t *testing.T) {
	client, bucket := mustMinio(t)
	viper.Reset()
	viper.Set(config.MinioBucketName, bucket)

	svc := NewMinioService(client)

	ctx := context.Background()
	objectName := fmt.Sprintf("test/%s.txt", uuid.NewString())
	payload := []byte("hello")

	if err := svc.UploadObject(ctx, objectName, bytesReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("UploadObject() error = %v", err)
	}

	if _, err := client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{}); err != nil {
		t.Fatalf("StatObject() error = %v", err)
	}

	if svc.GetObjectURL(objectName) == "" {
		t.Fatal("GetObjectURL() empty")
	}
	if svc.GetBaseURL() == "" {
		t.Fatal("GetBaseURL() empty")
	}

	if err := svc.DeleteObject(ctx, objectName); err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}
}

func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }

func TestMapUserWriteErrorBehavior(t *testing.T) {
	// Arrange
	originalErr := errors.New("some error")

	// Case 1: mapUserWriteError returns same error
	t.Run("same error returned", func(t *testing.T) {
		mapped := mapUserWriteError(originalErr)

		// Current logic: mappedErr != err
		usingEquality := mapped != originalErr

		// Proposed logic: !errors.Is(mappedErr, err)
		usingErrorsIs := !errors.Is(mapped, originalErr)

		t.Logf("Equality check: %v", usingEquality)
		t.Logf("errors.Is check: %v", usingErrorsIs)

		if usingEquality != usingErrorsIs {
			t.Errorf("Behavior differs! Equality=%v, errors.Is=%v",
				usingEquality, usingErrorsIs)
		}
	})

	// Case 2: mapUserWriteError returns domain error
	t.Run("domain error returned", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505"}
		mapped := mapUserWriteError(pgErr)

		usingEquality := mapped != pgErr
		usingErrorsIs := !errors.Is(mapped, pgErr)

		t.Logf("Equality check: %v", usingEquality)
		t.Logf("errors.Is check: %v", usingErrorsIs)

		if usingEquality != usingErrorsIs {
			t.Errorf("Behavior differs! Equality=%v, errors.Is=%v",
				usingEquality, usingErrorsIs)
		}
	})

	// Case 3: if mapUserWriteError wraps (hypothetical)
	t.Run("wrapped error", func(t *testing.T) {
		originalErr = &pgconn.PgError{Code: "23505"}
		// Simulate if mapUserWriteError starts wrapping
		wrappedErr := fmt.Errorf("wrapped: %w", originalErr)

		usingEquality := wrappedErr != originalErr
		usingErrorsIs := !errors.Is(wrappedErr, originalErr)

		t.Logf("Equality check: %v (expects true)", usingEquality)
		t.Logf("errors.Is check: %v (expects false!)", usingErrorsIs)

		// This shows the difference
		if usingEquality == usingErrorsIs {
			t.Logf("Both methods agree")
		} else {
			t.Logf("DIFFERENCE DETECTED: Equality=%v, errors.Is=%v",
				usingEquality, usingErrorsIs)
		}
	})
}
