package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/luis13005/ratelimiter/internal/config"
	"github.com/luis13005/ratelimiter/internal/storage"
)

type Limiter struct {
	Storage           storage.RedisInterface
	IpLimitRps        int
	IpBlockSeconds    int
	TokenLimitsRps    int
	TokenBlockSeconds int
}

type AllowResult struct {
	Allowed    bool
	Limit      int
	Current    int64
	IsToken    bool
	RetryAfter time.Duration
}

func NewLimiter(storage storage.RedisInterface, config *config.Conf) *Limiter {
	return &Limiter{
		Storage:           storage,
		IpLimitRps:        config.IpLimitRps,
		IpBlockSeconds:    int(config.IpBlockDuration.Seconds()),
		TokenLimitsRps:    config.TokenLimitRps,
		TokenBlockSeconds: int(config.TokenBlockDuration.Seconds()),
	}

}

func (l *Limiter) Allow(ctx context.Context, key string, isToken bool) (*AllowResult, error) {
	limit, blockSeconds := l.resolvePolicy(isToken)

	redisKey := buildKey(key, isToken)

	blocked, err := l.Storage.IsBlocked(ctx, redisKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar bloqueio: %w", err)
	}

	if blocked {
		ttl, err := l.Storage.TimeToRestore(ctx, redisKey) // ← busca o tempo restante
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar TTL: %w", err)
		}
		return &AllowResult{Allowed: false,
			Limit:   limit,
			Current: int64(limit) + 1, IsToken: isToken,
			RetryAfter: ttl}, nil
	}

	current, err := l.Storage.Increment(ctx, redisKey, blockSeconds)
	if err != nil {
		return nil, fmt.Errorf("erro ao incrementar contador: %w", err)
	}

	if current > int64(limit) {
		if err := l.Storage.Block(ctx, redisKey, blockSeconds); err != nil {
			return nil, fmt.Errorf("erro ao bloquear chave: %w", err)
		}
		return &AllowResult{Allowed: false,
			Limit:      limit,
			Current:    current,
			IsToken:    isToken,
			RetryAfter: time.Duration(blockSeconds) * time.Second}, nil
	}

	return &AllowResult{Allowed: true, Limit: limit, Current: current, IsToken: isToken}, nil
}

func (l *Limiter) resolvePolicy(isToken bool) (limit int, window int) {
	if isToken {
		return l.TokenLimitsRps, l.TokenBlockSeconds
	}
	return l.IpLimitRps, l.IpBlockSeconds
}

func buildKey(key string, isToken bool) string {
	if isToken {
		return "token:" + key
	}
	return "ip:" + key
}
