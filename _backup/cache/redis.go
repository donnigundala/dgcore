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

	"github.com/donnigundala/dgcore/ctxutil"
)

// redis implements the Provider interface for Redis.
type redis struct {
	client *goRedis.Client
	config *Config
	logger *slog.Logger // Base logger for the driver
}

// newRedisProvider creates and initializes a Redis client.
func newRedisProvider(cfg *Config) (Provider, error) {
	// Use the base logger provided by the manager/provider factory.
	logger := cfg.Logger.With("driver", "redis")

	options := &goRedis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Username: cfg.Redis.Username,
	}

	if cfg.Redis.EnableTLS {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, // Consider making this configurable for production
		}
	}

	rdb := goRedis.NewClient(options)
	// Ping with a background context as this is a connection-level check, not request-specific.
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
	// Use the context-aware logger for request-specific operations.
	logger := ctxutil.LoggerFromContext(ctx)
	logger.Debug("Pinging Redis")
	return r.client.Ping(ctx).Err()
}

// Set stores a key-value pair in Redis with optional expiration.
func (r *redis) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	logger := ctxutil.LoggerFromContext(ctx)
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
			logger.Error("failed to marshal value for Set operation", "key", key, "error", err)
			return fmt.Errorf("redis: failed to marshal value for key %s: %w", key, err)
		}
		data = b
	}

	err := r.client.Set(ctx, r.buildKey(key), data, expiration).Err()
	if err != nil {
		logger.Error("failed to set key in Redis", "key", key, "error", err)
	} else {
		logger.Debug("Set key in Redis", "key", key, "expiration", expiration)
	}
	return err
}

// Get retrieves a value by key and unmarshals it into the destination.
func (r *redis) Get(ctx context.Context, key string, dest any) error {
	logger := ctxutil.LoggerFromContext(ctx)
	val, err := r.client.Get(ctx, r.buildKey(key)).Result()
	if errors.Is(err, goRedis.Nil) {
		logger.Debug("Key not found in Redis", "key", key)
		return nil // key not found
	}
	if err != nil {
		logger.Warn("failed to get key from Redis", "key", key, "error", err)
		return fmt.Errorf("redis: failed to get key %s: %w", key, err)
	}

	switch d := dest.(type) {
	case *string:
		*d = val
	case *[]byte:
		*d = []byte(val)
	default:
		if err := json.Unmarshal([]byte(val), dest); err != nil {
			logger.Error("failed to unmarshal JSON from Redis", "key", key, "error", err)
			return fmt.Errorf("redis: failed to unmarshal JSON for key %s: %w", key, err)
		}
	}
	logger.Debug("Retrieved key from Redis", "key", key)
	return nil
}

// Delete removes a key from Redis.
func (r *redis) Delete(ctx context.Context, key string) error {
	logger := ctxutil.LoggerFromContext(ctx)
	err := r.client.Del(ctx, r.buildKey(key)).Err()
	if err != nil {
		logger.Error("failed to delete key from Redis", "key", key, "error", err)
	} else {
		logger.Debug("Deleted key from Redis", "key", key)
	}
	return err
}

// MGet retrieves multiple keys and unmarshals each value into the provided slice of destinations.
func (r *redis) MGet(ctx context.Context, keys []string, dests []any) error {
	logger := ctxutil.LoggerFromContext(ctx)
	if len(keys) != len(dests) {
		return fmt.Errorf("MGet: keys and destinations length mismatch")
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}

	vals, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		logger.Warn("failed to MGet keys from Redis", "keys", keys, "error", err)
		return err
	}

	for i, val := range vals {
		if val == nil {
			logger.Debug("Key not found in MGet", "key", keys[i])
			continue // missing key
		}

		switch d := dests[i].(type) {
		case *string:
			*d = val.(string)
		case *[]byte:
			*d = []byte(val.(string))
		default:
			if err := json.Unmarshal([]byte(val.(string)), dests[i]); err != nil {
				logger.Warn("failed to unmarshal value in MGet", "key", keys[i], "error", err)
				return fmt.Errorf("MGet: failed to unmarshal value for key %s: %w", keys[i], err)
			}
		}
	}
	logger.Debug("Performed MGet on Redis", "keys", keys)
	return nil
}

// MDelete removes multiple keys from Redis.
func (r *redis) MDelete(ctx context.Context, keys ...string) error {
	logger := ctxutil.LoggerFromContext(ctx)
	if len(keys) == 0 {
		return nil
	}
	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}
	err := r.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		logger.Error("failed to MDelete keys from Redis", "keys", keys, "error", err)
	} else {
		logger.Debug("Performed MDelete on Redis", "keys", keys)
	}
	return err
}

// ScanKeys scans and returns all keys matching a given pattern.
func (r *redis) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	var cursor uint64
	var keys []string
	fullPattern := r.buildKey(pattern)

	for {
		var res []string
		var err error
		res, cursor, err = r.client.Scan(ctx, cursor, fullPattern, limit).Result()
		if err != nil {
			logger.Error("failed to scan keys from Redis", "pattern", pattern, "error", err)
			return nil, err
		}
		keys = append(keys, res...)
		if cursor == 0 {
			break
		}
	}
	logger.Debug("Scanned keys from Redis", "pattern", pattern, "count", len(keys))
	return keys, nil
}

// Incr increments an integer value by 1.
func (r *redis) Incr(ctx context.Context, key string) (int64, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	val, err := r.client.Incr(ctx, r.buildKey(key)).Result()
	if err != nil {
		logger.Error("failed to increment key in Redis", "key", key, "error", err)
	} else {
		logger.Debug("Incremented key in Redis", "key", key, "value", val)
	}
	return val, err
}

// Decr decrements an integer value by 1.
func (r *redis) Decr(ctx context.Context, key string) (int64, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	val, err := r.client.Decr(ctx, r.buildKey(key)).Result()
	if err != nil {
		logger.Error("failed to decrement key in Redis", "key", key, "error", err)
	} else {
		logger.Debug("Decremented key in Redis", "key", key, "value", val)
	}
	return val, err
}

// Exists checks if one or more keys exist.
func (r *redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	if len(keys) == 0 {
		return 0, nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}

	count, err := r.client.Exists(ctx, fullKeys...).Result()
	if err != nil {
		logger.Error("failed to check existence of keys in Redis", "keys", keys, "error", err)
	} else {
		logger.Debug("Checked existence of keys in Redis", "keys", keys, "count", count)
	}
	return count, err
}

// Expire sets a TTL on a key.
func (r *redis) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	ok, err := r.client.Expire(ctx, r.buildKey(key), ttl).Result()
	if err != nil {
		logger.Error("failed to set expiration for key in Redis", "key", key, "ttl", ttl, "error", err)
	} else {
		logger.Debug("Set expiration for key in Redis", "key", key, "ttl", ttl, "success", ok)
	}
	return ok, err
}

// Ttl returns the time to live of a key.
func (r *redis) Ttl(ctx context.Context, key string) (time.Duration, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	duration, err := r.client.TTL(ctx, r.buildKey(key)).Result()
	if err != nil {
		if errors.Is(err, goRedis.Nil) {
			logger.Debug("TTL requested for non-existent key in Redis", "key", key)
			return 0, nil // Key does not exist
		}
		logger.Error("failed to get TTL for key from Redis", "key", key, "error", err)
		return 0, err
	}
	logger.Debug("Retrieved TTL for key from Redis", "key", key, "ttl", duration)
	return duration, nil
}

// Close gracefully closes the Redis connection.
func (r *redis) Close() error {
	// Use the base logger for connection-level operations.
	r.logger.Info("closing redis connection")
	return r.client.Close()
}
