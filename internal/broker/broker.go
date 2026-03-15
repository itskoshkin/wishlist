package broker

import (
	"context"
)

type Producer interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Close() error
}

type Handler func(ctx context.Context, msg []byte) error

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler Handler) error
	Close() error
}
