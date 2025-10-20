// Package dgredis Package redis implements a production-ready Redis cache client.
// It follows the same modular structure as pkg/database/mysql or pgsql,
// supporting config defaults, connection initialization, and graceful shutdown.
package dgredis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// -----------------------------
// Client
// -----------------------------

type Redis struct {
	client *redis.Client
	config *Config
}

// New creates and initializes a Redis client using the provided configuration.
func New(cfg *Config) *Redis {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	options := &redis.Options{
		Addr:     addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.EnableTLS {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	rdb := redis.NewClient(options)

	return &Redis{
		client: rdb,
		config: cfg,
	}
}

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
	return r.client.Set(ctx, r.buildKey(key), value, expiration).Err()
}

// Get retrieves a string value by key.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, r.buildKey(key)).Result()
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
	val, err := r.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil // Key not found
		}
		return err
	}
	return json.Unmarshal([]byte(val), dest)
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
	return r.client.Incr(ctx, r.buildKey(key)).Result()
}

// Decr decrements an integer value by 1.
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, r.buildKey(key)).Result()
}

// Exists checks if one or more keys exist.
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	var fullKeys []string
	for _, k := range keys {
		fullKeys = append(fullKeys, r.buildKey(k))
	}
	return r.client.Exists(ctx, fullKeys...).Result()
}

// Expire sets a TTL on a key.
func (r *Redis) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.Expire(ctx, r.buildKey(key), ttl).Result()
}

// Ttl returns the time to live of a key.
func (r *Redis) Ttl(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, r.buildKey(key)).Result()
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

// InitRedis initializes a Redis client, verifies connectivity, and logs the result.
// used for bootstrapping in main applications.
// just copy and paste into your main.go
//func InitRedis(cfg *Config) *Client {
//	client := New(cfg)
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := client.Ping(ctx); err != nil {
//		log.Fatalf("❌ Redis connection failed: %v", err)
//	}
//
//	log.Printf("✅ Connected to Redis at %s:%s (DB %d, namespace=%s)", cfg.Host, cfg.Port, cfg.DB, cfg.namespace)
//	return client
//}
