package cache

import (
	"context"
	"log/slog"
	"time"
)

// Provider defines the interface for cache operations.
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

// ProviderOption defines a functional option for configuring a cache provider.
type ProviderOption func(*Config)

// WithProviderLogger sets a base logger for the provider factory.
// The factory will create a sub-logger from this.
func WithProviderLogger(logger *slog.Logger) ProviderOption {
	return func(c *Config) {
		c.Logger = logger
	}
}
