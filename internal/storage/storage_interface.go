package storage

import (
	"context"
	"time"
)

type RedisInterface interface {
	Increment(ctx context.Context, key string, blockSeconds int) (int64, error)

	IsBlocked(ctx context.Context, key string) (bool, error)

	Block(ctx context.Context, key string, durationSeconds int) error

	TimeToRestore(ctx context.Context, key string) (time.Duration, error)
}
