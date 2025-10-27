package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// buildKey constructs the full key with namespace prefix.
func (r *Redis) buildKey(parts ...string) string {
	key := strings.Join(parts, r.config.Separator)
	if r.config.Namespace != "" {
		return fmt.Sprintf("%s%s%s", r.config.Namespace, r.config.Separator, key)
	}
	return key
}

// Ping verifies the Redis connection.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set stores a key-value pair in Redis with optional expiration.
func (r *Redis) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	expiration := r.config.TTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	switch v := value.(type) {
	case string, []byte:
		return r.client.Set(ctx, r.buildKey(key), v, expiration).Err()
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("redis: failed to marshal value for key %s: %w", key, err)
		}
		return r.client.Set(ctx, r.buildKey(key), b, expiration).Err()
	}
}

// Get retrieves a string value by key.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, r.buildKey(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // key not found
	}
	if err != nil {
		return "", fmt.Errorf("redis: failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Delete removes a key from Redis.
func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.buildKey(key)).Err()
}

// -----------------------------
// JSON Helpers
// -----------------------------

// SetJSON stores a Go value as JSON in Redis.
func (r *Redis) SetJSON(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.Set(ctx, key, string(data), ttl...)
}

// GetJSON retrieves and unmarshals JSON data from Redis into the provided destination.
func (r *Redis) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, r.buildKey(key)).Result()
	if errors.Is(err, redis.Nil) {
		return nil // key not found, return nil for consistency
	}
	if err != nil {
		return fmt.Errorf("redis: failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("redis: failed to unmarshal JSON for key %s: %w", key, err)
	}
	return nil
}

// -----------------------------
// Bulk Operations
// -----------------------------

// MGetJSON retrieves multiple keys and unmarshals each JSON value into the provided slice of destinations.
// The length of dests must match the number of keys.
func (r *Redis) MGetJSON(ctx context.Context, keys []string, dests []any) error {
	if len(keys) != len(dests) {
		return fmt.Errorf("MGetJSON: keys and destinations length mismatch")
	}

	var fullKeys []string
	for _, k := range keys {
		fullKeys = append(fullKeys, r.buildKey(k))
	}

	vals, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return err
	}

	for i, val := range vals {
		if val == nil {
			continue // missing key
		}
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("MGetJSON: expected string, got %T", val)
		}
		if err := json.Unmarshal([]byte(strVal), dests[i]); err != nil {
			return fmt.Errorf("MGetJSON: failed to unmarshal key %s: %w", keys[i], err)
		}
	}
	return nil
}

// MDelete removes multiple keys from Redis.
func (r *Redis) MDelete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	var fullKeys []string
	for _, k := range keys {
		fullKeys = append(fullKeys, r.buildKey(k))
	}
	return r.client.Del(ctx, fullKeys...).Err()
}

// ScanKeys scans and returns all keys matching a given pattern.
// Example: pattern = "user:*" or "cache:session:*"
func (r *Redis) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	var cursor uint64
	var keys []string
	fullPattern := r.buildKey(pattern)

	for {
		res, nextCursor, err := r.client.Scan(ctx, cursor, fullPattern, limit).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, res...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// -----------------------------
// Atomic Operations
// -----------------------------

// Incr increments an integer value by 1.
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, r.buildKey(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// Decr decrements an integer value by 1.
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Decr(ctx, r.buildKey(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: failed to decrement key %s: %w", key, err)
	}
	return val, nil
}

// Exists checks if one or more keys exist.
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	// Apply namespace
	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.buildKey(k)
	}

	count, err := r.client.Exists(ctx, fullKeys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: failed to check key existence: %w", err)
	}
	return count, nil
}

// Expire sets a TTL on a key.
func (r *Redis) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := r.client.Expire(ctx, r.buildKey(key), ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis: failed to set expiration for key %s: %w", key, err)
	}
	return ok, nil
}

// Ttl returns the time to live of a key.
func (r *Redis) Ttl(ctx context.Context, key string) (time.Duration, error) {
	duration, err := r.client.TTL(ctx, r.buildKey(key)).Result()
	if err != nil {
		return 0, err
	}

	switch {
	case duration == -2*time.Second:
		// Key does not exist
		return 0, nil
	case duration == -1*time.Second:
		// Key exists but has no expiration
		return -1, nil
	default:
		// Normal TTL
		return duration, nil
	}
}

// -----------------------------
// Close & Bootstrap
// -----------------------------

// Close gracefully closes the Redis connection.
func (r *Redis) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}
	return nil
}
