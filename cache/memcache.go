package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	goMemcache "github.com/bradfitz/gomemcache/memcache"
)

// memcache implements the Provider interface for Memcache.
type memcache struct {
	client *goMemcache.Client
	config *Config
	logger *slog.Logger
}

// newMemcache creates and initializes a Memcache client.
func newMemcache(cfg *Config) (Provider, error) {
	logger := cfg.Logger.With("driver", "memcache")

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	mc := goMemcache.New(addr)

	// Use Ping to verify the connection, as it's more direct.
	if err := mc.Ping(); err != nil {
		logger.Error("failed to connect to memcache", "error", err)
		return nil, fmt.Errorf("failed to connect to memcache: %w", err)
	}

	logger.Info("successfully connected to memcache", "host", cfg.Host, "port", cfg.Port)

	return &memcache{client: mc, config: cfg, logger: logger}, nil
}

// buildKey constructs the full key with namespace prefix.
func (m *memcache) buildKey(parts ...string) string {
	key := strings.Join(parts, m.config.Separator)
	if m.config.Namespace != "" {
		return fmt.Sprintf("%s%s%s", m.config.Namespace, m.config.Separator, key)
	}
	return key
}

// Ping verifies the Memcache connection.
func (m *memcache) Ping(ctx context.Context) error {
	return m.client.Ping()
}

// Set stores a key-value pair in Memcache with an optional expiration duration.
func (m *memcache) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	expiration := m.config.TTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			m.logger.Error("failed to marshal value", "key", key, "error", err)
			return fmt.Errorf("memcache: failed to marshal value for key %s: %w", key, err)
		}
		data = b
	}

	item := &goMemcache.Item{
		Key:        m.buildKey(key),
		Value:      data,
		Expiration: int32(expiration.Seconds()),
	}
	return m.client.Set(item)
}

// Get retrieves a value by key and unmarshals it into the destination.
func (m *memcache) Get(ctx context.Context, key string, dest any) error {
	item, err := m.client.Get(m.buildKey(key))
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		return nil // key not found
	}
	if err != nil {
		m.logger.Warn("failed to get key", "key", key, "error", err)
		return fmt.Errorf("memcache: failed to get key %s: %w", key, err)
	}

	switch d := dest.(type) {
	case *string:
		*d = string(item.Value)
	case *[]byte:
		*d = item.Value
	default:
		if err := json.Unmarshal(item.Value, dest); err != nil {
			m.logger.Error("failed to unmarshal JSON", "key", key, "error", err)
			return fmt.Errorf("memcache: failed to unmarshal JSON for key %s: %w", key, err)
		}
	}
	return nil
}

// Delete removes a key from Memcache.
func (m *memcache) Delete(ctx context.Context, key string) error {
	err := m.client.Delete(m.buildKey(key))
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		return nil // Deleting a non-existent key is not an error.
	}
	return err
}

// MGet retrieves multiple keys and unmarshals each value into the provided slice of destinations.
func (m *memcache) MGet(ctx context.Context, keys []string, dests []any) error {
	if len(keys) != len(dests) {
		return fmt.Errorf("MGet: keys and destinations length mismatch")
	}

	for i, key := range keys {
		if err := m.Get(ctx, key, dests[i]); err != nil {
			m.logger.Warn("failed to get key in mget", "key", key, "error", err)
			return fmt.Errorf("MGet: failed on key '%s': %w", key, err)
		}
	}
	return nil
}

// MDelete deletes multiple keys from Memcache.
func (m *memcache) MDelete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		err := m.Delete(ctx, key) // Ignore errors for individual deletes
		if err != nil {
			m.logger.Warn("failed to delete key in mdelete", "key", key, "error", err)
		}
	}
	return nil
}

// ScanKeys is not supported in Memcache.
func (m *memcache) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	return nil, errors.New("memcache: ScanKeys is not supported")
}

// Incr increments an integer value by 1.
func (m *memcache) Incr(ctx context.Context, key string) (int64, error) {
	fullKey := m.buildKey(key)
	newVal, err := m.client.Increment(fullKey, 1)
	if err == nil {
		return int64(newVal), nil
	}

	if errors.Is(err, goMemcache.ErrCacheMiss) {
		item := &goMemcache.Item{
			Key:        fullKey,
			Value:      []byte("1"),
			Expiration: int32(m.config.TTL.Seconds()),
		}
		if err := m.client.Add(item); err != nil {
			m.logger.Error("failed to initialize key for incr", "key", key, "error", err)
			return 0, fmt.Errorf("memcache: failed to initialize key for incr %s: %w", key, err)
		}
		return 1, nil
	}

	return 0, fmt.Errorf("memcache: failed to increment key %s: %w", key, err)
}

// Decr decrements an integer value by 1.
func (m *memcache) Decr(ctx context.Context, key string) (int64, error) {
	fullKey := m.buildKey(key)
	newVal, err := m.client.Decrement(fullKey, 1)
	if err == nil {
		return int64(newVal), nil
	}

	if errors.Is(err, goMemcache.ErrCacheMiss) {
		item := &goMemcache.Item{
			Key:        fullKey,
			Value:      []byte("0"), // Decrementing a non-existent key results in 0
			Expiration: int32(m.config.TTL.Seconds()),
		}
		if err := m.client.Add(item); err != nil {
			m.logger.Error("failed to initialize key for decr", "key", key, "error", err)
			return 0, fmt.Errorf("memcache: failed to initialize key for decr %s: %w", key, err)
		}
		return 0, nil
	}

	return 0, fmt.Errorf("memcache: failed to decrement key %s: %w", key, err)
}

// Exists checks if one or more keys exist.
func (m *memcache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var count int64
	for _, key := range keys {
		_, err := m.client.Get(m.buildKey(key))
		if err == nil {
			count++
		} else if !errors.Is(err, goMemcache.ErrCacheMiss) {
			m.logger.Warn("failed to check existence of key", "key", key, "error", err)
			return count, fmt.Errorf("memcache: failed to check existence of key %s: %w", key, err)
		}
	}
	return count, nil
}

// Expire sets a new expiration for a key.
func (m *memcache) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	err := m.client.Touch(m.buildKey(key), int32(ttl.Seconds()))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		return false, nil // Key doesn't exist
	}
	return false, fmt.Errorf("memcache: failed to update expiration for key %s: %w", key, err)
}

// Ttl is not supported in Memcache and is emulated.
func (m *memcache) Ttl(ctx context.Context, key string) (time.Duration, error) {
	_, err := m.client.Get(m.buildKey(key))
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		return 0, nil // Key does not exist or is expired.
	}
	if err != nil {
		return 0, fmt.Errorf("memcache: failed to get key %s for ttl check: %w", key, err)
	}
	// Memcache doesn't support querying TTL. -1 indicates "exists but TTL unknown".
	return -1, nil
}

// Close closes the Memcache client connection.
func (m *memcache) Close() error {
	m.logger.Info("closing memcache connection")
	// The underlying client from bradfitz/gomemcache doesn't have a Close method.
	return nil
}
