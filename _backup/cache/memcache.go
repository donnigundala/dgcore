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

	"github.com/donnigundala/dgcore/ctxutil"
)

// memcache implements the Provider interface for Memcache.
type memcache struct {
	client *goMemcache.Client
	config *Config
	logger *slog.Logger // Base logger for the driver
}

// newMemcacheProvider creates and initializes a Memcache client.
func newMemcacheProvider(cfg *Config) (Provider, error) {
	// Use the base logger provided by the manager/provider factory.
	logger := cfg.Logger.With("driver", "memcache")

	if cfg.Memcache == nil || len(cfg.Memcache.Servers) == 0 {
		return nil, errors.New("memcache configuration is missing or has no servers defined")
	}

	// The gomemcache client accepts one or more server addresses.
	mc := goMemcache.New(cfg.Memcache.Servers...)

	// Use Ping to verify the connection.
	if err := mc.Ping(); err != nil {
		logger.Error("failed to connect to memcache", "servers", cfg.Memcache.Servers, "error", err)
		return nil, fmt.Errorf("failed to connect to memcache servers %v: %w", cfg.Memcache.Servers, err)
	}

	logger.Info("successfully connected to memcache", "servers", cfg.Memcache.Servers)

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
	logger := ctxutil.LoggerFromContext(ctx)
	logger.Debug("Pinging Memcache")
	return m.client.Ping()
}

// Set stores a key-value pair in Memcache with an optional expiration duration.
func (m *memcache) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	logger := ctxutil.LoggerFromContext(ctx)
	expiration := m.config.Memcache.TTL
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
			logger.Error("failed to marshal value for Set operation", "key", key, "error", err)
			return fmt.Errorf("memcache: failed to marshal value for key %s: %w", key, err)
		}
		data = b
	}

	item := &goMemcache.Item{
		Key:        m.buildKey(key),
		Value:      data,
		Expiration: int32(expiration.Seconds()),
	}
	err := m.client.Set(item)
	if err != nil {
		logger.Error("failed to set key in Memcache", "key", key, "error", err)
	} else {
		logger.Debug("Set key in Memcache", "key", key, "expiration", expiration)
	}
	return err
}

// Get retrieves a value by key and unmarshals it into the destination.
func (m *memcache) Get(ctx context.Context, key string, dest any) error {
	logger := ctxutil.LoggerFromContext(ctx)
	item, err := m.client.Get(m.buildKey(key))
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		logger.Debug("Key not found in Memcache", "key", key)
		return nil // key not found
	}
	if err != nil {
		logger.Warn("failed to get key from Memcache", "key", key, "error", err)
		return fmt.Errorf("memcache: failed to get key %s: %w", key, err)
	}

	switch d := dest.(type) {
	case *string:
		*d = string(item.Value)
	case *[]byte:
		*d = item.Value
	default:
		if err := json.Unmarshal(item.Value, dest); err != nil {
			logger.Error("failed to unmarshal JSON from Memcache", "key", key, "error", err)
			return fmt.Errorf("memcache: failed to unmarshal JSON for key %s: %w", key, err)
		}
	}
	logger.Debug("Retrieved key from Memcache", "key", key)
	return nil
}

// Delete removes a key from Memcache.
func (m *memcache) Delete(ctx context.Context, key string) error {
	logger := ctxutil.LoggerFromContext(ctx)
	err := m.client.Delete(m.buildKey(key))
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		logger.Debug("Attempted to delete non-existent key from Memcache", "key", key)
		return nil // Deleting a non-existent key is not an error.
	}
	if err != nil {
		logger.Error("failed to delete key from Memcache", "key", key, "error", err)
	} else {
		logger.Debug("Deleted key from Memcache", "key", key)
	}
	return err
}

// MGet retrieves multiple keys and unmarshals each value into the provided slice of destinations.
func (m *memcache) MGet(ctx context.Context, keys []string, dests []any) error {
	logger := ctxutil.LoggerFromContext(ctx)
	if len(keys) != len(dests) {
		return fmt.Errorf("MGet: keys and destinations length mismatch")
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = m.buildKey(k)
	}

	items, err := m.client.GetMulti(fullKeys)
	if err != nil {
		logger.Warn("failed to MGet keys from Memcache", "keys", keys, "error", err)
		return err
	}

	for i, key := range keys {
		fullKey := m.buildKey(key)
		item, ok := items[fullKey]
		if !ok {
			logger.Debug("Key not found in MGet", "key", key)
			continue // key not found
		}

		switch d := dests[i].(type) {
		case *string:
			*d = string(item.Value)
		case *[]byte:
			*d = item.Value
		default:
			if err := json.Unmarshal(item.Value, dests[i]); err != nil {
				logger.Error("failed to unmarshal JSON in MGet", "key", key, "error", err)
				return fmt.Errorf("memcache: failed to unmarshal JSON for key %s: %w", key, err)
			}
		}
	}
	logger.Debug("Performed MGet on Memcache", "keys", keys)
	return nil
}

// MDelete deletes multiple keys from Memcache.
func (m *memcache) MDelete(ctx context.Context, keys ...string) error {
	logger := ctxutil.LoggerFromContext(ctx)
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		err := m.Delete(ctx, key) // Delete uses the context-aware logger internally.
		if err != nil {
			logger.Warn("failed to delete key in MDelete", "key", key, "error", err)
		}
	}
	logger.Debug("Performed MDelete on Memcache", "keys", keys)
	return nil
}

// ScanKeys is not supported in Memcache.
func (m *memcache) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	logger.Warn("ScanKeys is not supported by Memcache driver")
	return nil, errors.New("memcache: ScanKeys is not supported")
}

// Incr increments an integer value by 1.
func (m *memcache) Incr(ctx context.Context, key string) (int64, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	fullKey := m.buildKey(key)
	newVal, err := m.client.Increment(fullKey, 1)
	if err == nil {
		logger.Debug("Incremented key in Memcache", "key", key, "value", newVal)
		return int64(newVal), nil
	}

	if errors.Is(err, goMemcache.ErrCacheMiss) {
		item := &goMemcache.Item{
			Key:        fullKey,
			Value:      []byte("1"),
			Expiration: int32(m.config.Memcache.TTL.Seconds()),
		}
		if err := m.client.Add(item); err != nil {
			logger.Error("failed to initialize key for Incr in Memcache", "key", key, "error", err)
			return 0, fmt.Errorf("memcache: failed to initialize key for Incr %s: %w", key, err)
		}
		logger.Debug("Initialized and incremented key in Memcache", "key", key, "value", 1)
		return 1, nil
	}

	logger.Error("failed to increment key in Memcache", "key", key, "error", err)
	return 0, fmt.Errorf("memcache: failed to increment key %s: %w", key, err)
}

// Decr decrements an integer value by 1.
func (m *memcache) Decr(ctx context.Context, key string) (int64, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	fullKey := m.buildKey(key)
	newVal, err := m.client.Decrement(fullKey, 1)
	if err == nil {
		logger.Debug("Decremented key in Memcache", "key", key, "value", newVal)
		return int64(newVal), nil
	}

	if errors.Is(err, goMemcache.ErrCacheMiss) {
		item := &goMemcache.Item{
			Key:        fullKey,
			Value:      []byte("0"), // Decrementing a non-existent key results in 0
			Expiration: int32(m.config.Memcache.TTL.Seconds()),
		}
		if err := m.client.Add(item); err != nil {
			logger.Error("failed to initialize key for Decr in Memcache", "key", key, "error", err)
			return 0, fmt.Errorf("memcache: failed to initialize key for Decr %s: %w", key, err)
		}
		logger.Debug("Initialized and decremented key in Memcache", "key", key, "value", 0)
		return 0, nil
	}

	logger.Error("failed to decrement key in Memcache", "key", key, "error", err)
	return 0, fmt.Errorf("memcache: failed to decrement key %s: %w", key, err)
}

// Exists checks if one or more keys exist.
func (m *memcache) Exists(ctx context.Context, keys ...string) (int64, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	if len(keys) == 0 {
		return 0, nil
	}

	var count int64
	for _, key := range keys {
		_, err := m.client.Get(m.buildKey(key))
		if err == nil {
			count++
		} else if !errors.Is(err, goMemcache.ErrCacheMiss) {
			logger.Warn("failed to check existence of key in Memcache", "key", key, "error", err)
			return count, fmt.Errorf("memcache: failed to check existence of key %s: %w", key, err)
		}
	}
	logger.Debug("Checked existence of keys in Memcache", "keys", keys, "count", count)
	return count, nil
}

// Expire sets a new expiration for a key.
func (m *memcache) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	err := m.client.Touch(m.buildKey(key), int32(ttl.Seconds()))
	if err == nil {
		logger.Debug("Set expiration for key in Memcache", "key", key, "ttl", ttl)
		return true, nil
	}
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		logger.Debug("Attempted to set expiration for non-existent key in Memcache", "key", key)
		return false, nil // Key doesn't exist
	}
	logger.Error("failed to update expiration for key in Memcache", "key", key, "error", err)
	return false, fmt.Errorf("memcache: failed to update expiration for key %s: %w", key, err)
}

// Ttl is not directly supported in Memcache and is emulated.
func (m *memcache) Ttl(ctx context.Context, key string) (time.Duration, error) {
	logger := ctxutil.LoggerFromContext(ctx)
	_, err := m.client.Get(m.buildKey(key))
	if errors.Is(err, goMemcache.ErrCacheMiss) {
		logger.Debug("TTL requested for non-existent key in Memcache", "key", key)
		return 0, nil // Key does not exist or is expired.
	}
	if err != nil {
		logger.Error("failed to get key for TTL check in Memcache", "key", key, "error", err)
		return 0, fmt.Errorf("memcache: failed to get key %s for TTL check: %w", key, err)
	}
	// Memcache doesn't support querying TTL directly. -1 indicates "exists but TTL unknown".
	logger.Debug("Retrieved TTL for key from Memcache (emulated)", "key", key, "ttl", -1)
	return -1, nil
}

// Close closes the Memcache client connection.
func (m *memcache) Close() error {
	// Use the base logger for connection-level operations.
	m.logger.Info("closing memcache connection")
	// The underlying client from bradfitz/gomemcache doesn't have a Close method.
	return nil
}
