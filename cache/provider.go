package cache

import (
	"context"
	"log/slog"
	"time"
)

// Provider defines the interface for cache operations.
//
// Note: Some methods are emulated in Memcache where not natively supported.
// - ScanKeys: not supported (returns error)
// - Ttl: not supported (returns error)
// - MGet, MDelete, Exists: emulated via loops
type Provider interface {
	Ping(ctx context.Context) error
	Set(ctx context.Context, key string, value any, ttl ...time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	MGet(ctx context.Context, keys []string, dests []any) error
	MDelete(ctx context.Context, keys ...string) error
	ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Ttl(ctx context.Context, key string) (time.Duration, error)
	Close() error
}

// Driver defines the type for cache driver names.
type Driver string

var (
	DriverRedis    Driver = "redis"
	DriverMemcache Driver = "memcache"
)

// Option defines a functional option for configuring the cache.
type Option func(*Config)

// WithLogger sets a custom logger for the cache.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}
