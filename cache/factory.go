package cache

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// newProvider acts as an internal factory for creating a cache Provider.
// It is called by the Manager.
func newProvider(cfg *Config, opts ...ProviderOption) (Provider, error) {
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
