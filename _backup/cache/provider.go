package cache

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
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

// providerOption defines a functional option for configuring a cache provider.
type providerOption func(*Config)

// withProviderLogger sets a base logger for the provider factory.
// The factory will create a sub-logger from this.
func withProviderLogger(logger *slog.Logger) providerOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

// newProvider acts as an internal factory for creating a cache Provider.
// It is called by the Manager.
func newProvider(cfg *Config, opts ...providerOption) (Provider, error) {
	// Apply all the functional options to the config
	for _, opt := range opts {
		opt(cfg)
	}

	// If no logger was provided by the options (e.g., from the manager),
	// create a default one.
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	switch Driver(strings.ToLower(string(cfg.Driver))) {
	case DriverRedis:
		return newRedisProvider(cfg)
	case DriverMemcache:
		return newMemcacheProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", cfg.Driver)
	}
}
