package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"wishlist/internal/utils/colors"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	LogLevel string
}

func NewInstance(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	fmt.Printf("Connecting to Postgres...")

	config, err := pgxpool.ParseConfig(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode))
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("pgx: failed to parse config: %w", err)
	}

	config.MinConns = 2
	config.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("pgx: failed to create database connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		fmt.Println()
		return nil, fmt.Errorf("pgx: failed to connect to database: %w", err)
	}

	fmt.Println(colors.Green(" Done."))
	return pool, nil
}
