package memcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// buildKey constructs the full key with namespace prefix.
func (m *Memcache) buildKey(parts ...string) string {
	key := strings.Join(parts, m.config.Separator)
	if m.config.Namespace != "" {
		return fmt.Sprintf("%s%s%s", m.config.Namespace, m.config.Separator, key)
	}
	return key
}

// Ping verifies the Memcache connection.
func (m *Memcache) Ping(ctx context.Context) error {
	testKey := m.buildKey("__ping_test__")
	testValue := []byte("ok")

	// Set a short-lived dummy value
	err := m.client.Set(&memcache.Item{
		Key:        testKey,
		Value:      testValue,
		Expiration: int32(2),
	})
	if err != nil {
		return fmt.Errorf("memcache ping set failed: %w", err)
	}

	_, err = m.client.Get(testKey)
	if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return fmt.Errorf("memcache ping get failed: %w", err)
	}

	return nil
}

// Set stores a key-value pair in Memcache with an optional expiration duration
func (m *Memcache) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	expiration := m.config.TTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	// Prepare value as bytes
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("memcache: failed to marshal value for key %s: %w", key, err)
		}
		bytes = b
	}

	item := &memcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: int32(expiration.Seconds()), // 0 means no expiry
	}
	return m.client.Set(item)
}

// Get retrieves a value by key from Memcache
func (m *Memcache) Get(ctx context.Context, key string) (string, error) {
	item, err := m.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return "", nil // key not found
	}
	if err != nil {
		return "", fmt.Errorf("memcache: failed to get key %s: %w", key, err)
	}

	return string(item.Value), nil
}

// Delete removes a key from Memcache
func (m *Memcache) Delete(ctx context.Context, key string) error {
	return m.client.Delete(key)
}

// -----------------------------
// JSON Helpers
// -----------------------------

// SetJSON GetJSON retrieves and unmarshals JSON data from Memcache into the provided destination.
func (m *Memcache) SetJSON(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("memcache setjson marshal error: %w", err)
	}

	duration := m.config.TTL
	if len(ttl) > 0 && ttl[0] > 0 {
		duration = ttl[0]
	}

	item := &memcache.Item{
		Key:        m.buildKey(key),
		Value:      data,
		Expiration: int32(duration.Seconds()),
	}

	return m.client.Set(item)
}

// GetJSON retrieves and unmarshals JSON data from Memcache into the provided destination.
func (m *Memcache) GetJSON(ctx context.Context, key string, dest any) error {
	item, err := m.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil // key not found
	}
	if err != nil {
		return fmt.Errorf("memcache: failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(item.Value, dest); err != nil {
		return fmt.Errorf("memcache: failed to unmarshal JSON for key %s: %w", key, err)
	}
	return nil
}

// -----------------------------
// Bulk Operations
// -----------------------------

// MGetJSON retrieves multiple JSON values and unmarshals them into dests
// The length of dests must match the number of keys.
func (m *Memcache) MGetJSON(ctx context.Context, keys []string, dests []any) error {
	for i, key := range keys {
		if err := m.GetJSON(ctx, key, dests[i]); err != nil {
			return err
		}
	}
	return nil
}

// MDelete deletes multiple keys from Memcache
func (m *Memcache) MDelete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		_ = m.Delete(ctx, key)
	}
	return nil
}

// ScanKeys is not supported in Memcache
// Example: pattern = "user:*" or "cache:session:*"
func (r *Memcache) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	return nil, fmt.Errorf("ScanKeys not supported in Memcache")
}

// -----------------------------
// Atomic Operations
// -----------------------------

// Incr Expire sets a new expiration for a key
func (m *Memcache) Incr(ctx context.Context, key string) (int64, error) {
	// Try to increment
	newVal, err := m.client.Increment(key, 1)
	if err == nil {
		return int64(newVal), nil
	}

	// If key doesn't exist, create it with initial value 1
	if errors.Is(err, memcache.ErrCacheMiss) {
		item := &memcache.Item{
			Key:        key,
			Value:      []byte("1"),
			Expiration: int32(m.config.TTL.Seconds()),
		}
		if err := m.client.Add(item); err != nil {
			return 0, fmt.Errorf("memcache: failed to create key %s: %w", key, err)
		}
		return 1, nil
	}

	return 0, fmt.Errorf("memcache: failed to increment key %s: %w", key, err)
}

// Decr decrements the integer value of a key by one
func (m *Memcache) Decr(ctx context.Context, key string) (int64, error) {
	newVal, err := m.client.Decrement(key, 1)
	if err == nil {
		return int64(newVal), nil
	}

	if errors.Is(err, memcache.ErrCacheMiss) {
		// Key doesn't exist → initialize as 0 (so after decrement = 0, like Memcache spec)
		item := &memcache.Item{
			Key:        key,
			Value:      []byte("0"),
			Expiration: int32(m.config.TTL.Seconds()),
		}
		if err := m.client.Add(item); err != nil {
			return 0, fmt.Errorf("memcache: failed to create key %s: %w", key, err)
		}
		return 0, nil
	}

	return 0, fmt.Errorf("memcache: failed to decrement key %s: %w", key, err)
}

// Exists checks the existence of multiple keys and returns the count of existing keys
func (m *Memcache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var count int64
	for _, key := range keys {
		_, err := m.client.Get(key)
		if err == nil {
			count++
			continue
		}
		if err != memcache.ErrCacheMiss {
			return count, fmt.Errorf("memcache: failed to check existence of key %s: %w", key, err)
		}
	}
	return count, nil
}

// Expire sets a new expiration for a key
func (m *Memcache) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	item, err := m.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		// Key doesn't exist
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("memcache: failed to get key %s: %w", key, err)
	}

	// Re-set the same value with new TTL
	item.Expiration = int32(ttl.Seconds())
	if err := m.client.Replace(item); err != nil {
		return false, fmt.Errorf("memcache: failed to update expiration for key %s: %w", key, err)
	}
	return true, nil
}

// Ttl is not supported in Memcache
func (m *Memcache) Ttl(ctx context.Context, key string) (time.Duration, error) {
	_, err := m.client.Get(m.buildKey(key))
	if errors.Is(err, memcache.ErrCacheMiss) {
		return 0, nil // key expired or not found
	}
	if err != nil {
		return 0, err
	}
	// Memcache doesn't support TTL query; return -1 for "exists but TTL unknown"
	return -1, nil
}

// -----------------------------
// Close & Bootstrap
// -----------------------------

// Close closes the Memcache client connection
func (m *Memcache) Close() error {
	// memcache has no explicit close method
	return nil
}
