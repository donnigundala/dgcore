package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/donnigundala/dgcore/cache/provider"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	config *provider.Config
}

// New creates and initializes a Redis client using the provided configuration.
func New(cfg *provider.Config) (*Redis, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	options := &redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	// Set username if provided
	if cfg.Username != "" {
		options.Username = cfg.Username
	}

	// Enable TLS if specified
	if cfg.EnableTLS {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Create Redis client
	rdb := redis.NewClient(options)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Return Redis instance
	return &Redis{client: rdb, config: cfg}, nil
}
