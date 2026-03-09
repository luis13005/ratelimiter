package storage

import (
	"context"
	"time"

	"github.com/luis13005/ratelimiter/internal/config"
	"github.com/redis/go-redis/v9"
)

type ClientRedis struct {
	client *redis.Client
}

func NewClientRedis(conf *config.Conf) *ClientRedis {
	return &ClientRedis{
		client: redis.NewClient(&redis.Options{
			Addr:     conf.RedisAddr,
			Password: conf.RedisPassword,
			DB:       conf.RedisDB,
		}),
	}
}

func (r *ClientRedis) Increment(ctx context.Context, key string, windowSeconds int) (int64, error) {
	script := redis.NewScript(`
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then
			redis.call("EXPIRE", KEYS[1], ARGV[1])
		end
		return current
	`)

	result, err := script.Run(ctx, r.client, []string{key}, windowSeconds).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *ClientRedis) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := "blocked:" + key
	val, err := r.client.Exists(ctx, blockKey).Result()

	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (r *ClientRedis) Block(ctx context.Context, key string, durationSeconds int) error {
	blockKey := "blocked:" + key
	return r.client.Set(ctx, blockKey, 1, time.Duration(durationSeconds)*time.Second).Err()
}

func (r *ClientRedis) TimeToRestore(ctx context.Context, key string) (time.Duration, error) {
	blockKey := "blocked:" + key

	ttl, err := r.client.TTL(ctx, blockKey).Result()
	if err != nil {
		return 0, err
	}

	return ttl, nil
}
