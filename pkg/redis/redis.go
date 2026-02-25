package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/logging"

	"wishlist/internal/utils/colors"
)

type Config struct {
	Addr     string
	Port     string
	Password string
	Database int
}

func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	fmt.Printf("Connecting to Redis...")

	redis.SetLogger(&logging.VoidLogger{})
	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	_, err := rc.Ping(ctx).Result()
	if err != nil {
		fmt.Println()
		_ = rc.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	fmt.Println(colors.Green("    Done."))
	return rc, nil
}
