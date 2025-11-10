package cache

import (
	"fmt"
	"strings"
)

// New acts as an internal factory for creating a cache Provider.
// It is called by the CacheManager.
func New(cfg *Config, opts ...Option) (Provider, error) {
	// Apply all the functional options to the config
	for _, opt := range opts {
		opt(cfg)
	}

	// If no logger was provided, create a default one.
	// The logger from the manager is passed via WithLogger option.

	switch Driver(strings.ToLower(string(cfg.Driver))) {
	case DriverRedis:
		return newRedisProvider(cfg)
	case DriverMemcache:
		return newMemcacheProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", cfg.Driver)
	}
}
