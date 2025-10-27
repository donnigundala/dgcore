package cache

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// New returns a cache Provider based on the given config and options.
func New(cfg *Config, opts ...Option) (Provider, error) {
	// Apply all the functional options to the config
	for _, opt := range opts {
		opt(cfg)
	}

	// If no logger was provided, create a default one.
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// Create a sub-logger for the cache component.
	cfg.Logger = cfg.Logger.With("component", "cache")

	// set default to redis
	if cfg.Driver == "" {
		cfg.Driver = "redis"
	}

	switch Driver(strings.ToLower(string(cfg.Driver))) {
	case DriverRedis:
		return newRedis(cfg)
	case DriverMemcache:
		return newMemcache(cfg)
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", cfg.Driver)
	}
}
