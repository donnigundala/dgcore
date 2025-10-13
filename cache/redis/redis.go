// Package redis implements a production-ready Redis cache client.
// It follows the same modular structure as pkg/database/mysql or pgsql,
// supporting config defaults, connection initialization, and graceful shutdown.
package redis

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

type Client struct {
	Conn      *redis.Client
	TTL       time.Duration
	Namespace string
	Separator string
}

// NewRedis creates and initializes a Redis client using the provided configuration.
func NewRedis(cfg *Config) *Client {
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

	return &Client{
		Conn:      rdb,
		TTL:       cfg.TTL,
		Namespace: cfg.Namespace,
		Separator: cfg.Separator,
	}
}

// buildKey constructs the full key with namespace prefix.
func (c *Client) buildKey(parts ...string) string {
	key := strings.Join(parts, c.Separator)
	if c.Namespace != "" {
		return fmt.Sprintf("%s%s%s", c.Namespace, c.Separator, key)
	}
	return key
}

// Ping verifies the Redis connection.
func (c *Client) Ping(ctx context.Context) error {
	return c.Conn.Ping(ctx).Err()
}

// Set stores a key-value pair in Redis with optional expiration.
func (c *Client) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	expiration := c.TTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}
	return c.Conn.Set(ctx, c.buildKey(key), value, expiration).Err()
}

// Get retrieves a string value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Conn.Get(ctx, c.buildKey(key)).Result()
}

// Delete removes a key from Redis.
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.Conn.Del(ctx, c.buildKey(key)).Err()
}

// -----------------------------
// JSON Helpers
// -----------------------------

// SetJSON stores a Go value as JSON in Redis.
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return c.Set(ctx, key, string(data), ttl...)
}

// GetJSON retrieves and unmarshals JSON data from Redis into the provided destination.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := c.Get(ctx, key)
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
func (c *Client) MGetJSON(ctx context.Context, keys []string, dests []any) error {
	if len(keys) != len(dests) {
		return fmt.Errorf("MGetJSON: keys and destinations length mismatch")
	}

	var fullKeys []string
	for _, k := range keys {
		fullKeys = append(fullKeys, c.buildKey(k))
	}

	vals, err := c.Conn.MGet(ctx, fullKeys...).Result()
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
func (c *Client) MDelete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	var fullKeys []string
	for _, k := range keys {
		fullKeys = append(fullKeys, c.buildKey(k))
	}
	return c.Conn.Del(ctx, fullKeys...).Err()
}

// ScanKeys scans and returns all keys matching a given pattern.
// Example: pattern = "user:*" or "cache:session:*"
func (c *Client) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	var cursor uint64
	var keys []string
	fullPattern := c.buildKey(pattern)

	for {
		res, nextCursor, err := c.Conn.Scan(ctx, cursor, fullPattern, limit).Result()
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
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.Conn.Incr(ctx, c.buildKey(key)).Result()
}

// Decr decrements an integer value by 1.
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.Conn.Decr(ctx, c.buildKey(key)).Result()
}

// Exists checks if one or more keys exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	var fullKeys []string
	for _, k := range keys {
		fullKeys = append(fullKeys, c.buildKey(k))
	}
	return c.Conn.Exists(ctx, fullKeys...).Result()
}

// Expire sets a TTL on a key.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.Conn.Expire(ctx, c.buildKey(key), ttl).Result()
}

// Ttl returns the time to live of a key.
func (c *Client) Ttl(ctx context.Context, key string) (time.Duration, error) {
	return c.Conn.TTL(ctx, c.buildKey(key)).Result()
}

// -----------------------------
// Close & Bootstrap
// -----------------------------

// Close gracefully closes the Redis connection.
func (c *Client) Close() error {
	if err := c.Conn.Close(); err != nil {
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
//	log.Printf("✅ Connected to Redis at %s:%s (DB %d, namespace=%s)", cfg.Host, cfg.Port, cfg.DB, cfg.Namespace)
//	return client
//}
