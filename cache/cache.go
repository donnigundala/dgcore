package cache

import (
	"fmt"
	"strings"

	"github.com/donnigundala/dgcore/cache/memcache"
	"github.com/donnigundala/dgcore/cache/provider"
	"github.com/donnigundala/dgcore/cache/redis"
)

// New returns a cache Provider based on the given config.
func New(cfg *provider.Config) (provider.Provider, error) {
	// set default to redis
	if cfg.Driver == "" {
		cfg.Driver = "redis"
	}

	switch provider.Driver(strings.ToLower(string(cfg.Driver))) {
	case provider.DriverRedis:
		return redis.New(cfg)
	case provider.DriverMemcache:
		return memcache.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", cfg.Driver)
	}
}
