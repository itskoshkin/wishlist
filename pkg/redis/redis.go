package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Port     string
	Password string
	Database int
}

func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	_, err := rc.Ping(ctx).Result()
	if err != nil {
		_ = rc.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rc, nil
}
