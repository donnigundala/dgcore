package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

// redis implements the Provider interface for Redis.
type redis struct {
	client *goRedis.Client
	config *Config
	logger *slog.Logger
}

// newRedisProvider creates and initializes a Redis client.
func newRedisProvider(cfg *Config) (Provider, error) {
	logger := cfg.Logger.With("driver", "redis")

	options := &goRedis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Username: cfg.Redis.Username,
	}

	if cfg.Redis.EnableTLS {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, // Consider making this configurable
		}
	}

	rdb := goRedis.NewClient(options)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed to connect to redis", "error", err)
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("successfully connected to redis", "host", cfg.Redis.Host, "port", cfg.Redis.Port)

	return &redis{client: rdb, config: cfg, logger: logger}, nil
}

// buildKey constructs the full key with namespace prefix.
func (r *redis) buildKey(parts ...string) string {
	key := strings.Join(parts, r.config.Separator)
	if r.config.Namespace != "" {
		return fmt.Sprintf("%s%s%s", r.config.Namespace, r.config.Separator, key)
	}
	return key
}

// Ping verifies the Redis connection.
func (r *redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set stores a key-value pair in Redis with optional expiration.
func (r *redis) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	expiration := r.config.Redis.TTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	var data any
	switch v := value.(type) {
	case string, []byte:
		data = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			r.logger.Error("failed to marshal value", "key", key, "error", err)
			return fmt.Errorf("redis: failed to marshal value for key %s: %w", key, err)
		}
		data = b
	}

	return r.client.Set(ctx, r.buildKey(key), data, expiration).Err()
}

// Get retrieves a value by key and unmarshals it into the destination.
func (r *redis) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, r.buildKey(key)).Result()
	if errors.Is(err, goRedis.Nil) {
		return nil // key not found
	}
	if err != nil {
		r.logger.Warn("failed to get key", "key", key, "error", err)
		return fmt.Errorf("redis: failed to get key %s: %w", key, err)
	}

	switch d := dest.(type) {
	case *string:
		*d = val
	case *[]byte:
		*d = []byte(val)
	default:
		if err := json.Unmarshal([]byte(val), dest); err != nil {
			r.logger.Error("failed to unmarshal JSON", "key", key, "error", err)
			return fmt.Errorf("redis: failed to unmarshal JSON for key %s: %w", key, err)
		}
	}
	return nil
}

// Delete removes a key from Redis.
func (r *redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.buildKey(key)).Err()
}

// MGet retrieves multiple keys and unmarshals each value into the provided slice of destinations.
func (r *redis) MGet(ctx context.Context, keys []string, dests []any) error {
	if len(keys) != len(dests) {
		return fmt.Errorf("MGet: keys and destinations length mismatch")
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}

	vals, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		r.logger.Warn("failed to mget keys", "error", err)
		return err
	}

	for i, val := range vals {
		if val == nil {
			continue // missing key
		}

		switch d := dests[i].(type) {
		case *string:
			*d = val.(string)
		case *[]byte:
			*d = []byte(val.(string))
		default:
			if err := json.Unmarshal([]byte(val.(string)), dests[i]); err != nil {
				r.logger.Warn("failed to unmarshal value in mget", "key", keys[i], "error", err)
				return fmt.Errorf("MGet: failed to unmarshal value for key %s: %w", keys[i], err)
			}
		}
	}
	return nil
}

// MDelete removes multiple keys from Redis.
func (r *redis) MDelete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}
	return r.client.Del(ctx, fullKeys...).Err()
}

// ScanKeys scans and returns all keys matching a given pattern.
func (r *redis) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	var cursor uint64
	var keys []string
	fullPattern := r.buildKey(pattern)

	for {
		var res []string
		var err error
		res, cursor, err = r.client.Scan(ctx, cursor, fullPattern, limit).Result()
		if err != nil {
			r.logger.Error("failed to scan keys", "pattern", pattern, "error", err)
			return nil, err
		}
		keys = append(keys, res...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// Incr increments an integer value by 1.
func (r *redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, r.buildKey(key)).Result()
}

// Decr decrements an integer value by 1.
func (r *redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, r.buildKey(key)).Result()
}

// Exists checks if one or more keys exist.
func (r *redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}

	return r.client.Exists(ctx, fullKeys...).Result()
}

// Expire sets a TTL on a key.
func (r *redis) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.Expire(ctx, r.buildKey(key), ttl).Result()
}

// Ttl returns the time to live of a key.
func (r *redis) Ttl(ctx context.Context, key string) (time.Duration, error) {
	duration, err := r.client.TTL(ctx, r.buildKey(key)).Result()
	if err != nil {
		if errors.Is(err, goRedis.Nil) {
			return 0, nil // Key does not exist
		}
		return 0, err
	}

	return duration, nil
}

// Close gracefully closes the Redis connection.
func (r *redis) Close() error {
	r.logger.Info("closing redis connection")
	return r.client.Close()
}
