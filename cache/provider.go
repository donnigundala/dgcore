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

type Driver string

// Config defines Redis and Memcache connection and caching settings.
type Config struct {
	Driver Driver `mapstructure:"driver" json:"driver" yaml:"driver"` // dgredis

	// Host is the Redis server hostname or IP address.
	Host string `mapstructure:"host" json:"host" yaml:"host"`

	// Port is the Redis server port number.
	Port string `mapstructure:"port" json:"port" yaml:"port"`

	// Username for Redis authentication (if required).
	Username string `mapstructure:"username" json:"username" yaml:"username"`

	// Password for Redis authentication (if required).
	Password string `mapstructure:"password" json:"password" yaml:"password"`

	// DB is the Redis logical database index (0–15).
	// Defaults to 0. Override with REDIS_DB in environment variables.
	DB int `mapstructure:"db" json:"db" yaml:"db"`

	// TTL is the default time-to-live for cached items.
	TTL time.Duration `mapstructure:"ttl" json:"ttl" yaml:"ttl"`

	// EnableTLS indicates whether to use TLS for the Redis connection.
	EnableTLS bool `mapstructure:"enable_tls" json:"enable_tls" yaml:"enable_tls"`

	// Namespace is a prefix added to all keys to avoid collisions.
	Namespace string `mapstructure:"namespace" json:"namespace" yaml:"namespace"`

	// Separator is the string used to separate namespace and keys.
	Separator string `mapstructure:"separator" json:"separator" yaml:"separator"`

	// InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify" yaml:"insecure_skip_verify"`

	// Logger is the slog logger instance.
	Logger *slog.Logger `mapstructure:"-" json:"-" yaml:"-"`
}
