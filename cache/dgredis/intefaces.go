package dgredis

import (
	"context"
	"time"
)

// IRedis defines the interface for Redis operations.
type IRedis interface {
	Ping(ctx context.Context) error
	Set(ctx context.Context, key string, value any, ttl ...time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	SetJSON(ctx context.Context, key string, value any, ttl ...time.Duration) error
	GetJSON(ctx context.Context, key string, dest any) error
	MGetJSON(ctx context.Context, keys []string, dests []any) error
	MDelete(ctx context.Context, keys ...string) error
	ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Ttl(ctx context.Context, key string) (time.Duration, error)
	Close() error
}
