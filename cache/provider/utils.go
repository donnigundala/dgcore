package provider

import (
	"errors"
	"strings"
	"time"
)

// NormalizeTTLResult ensures consistent TTL behavior across drivers.
//
// Input:
//   - duration: raw TTL returned by the cache backend
//   - keyExists: whether the key exists
//
// Rules:
//   - If key doesn’t exist → (0, nil)
//   - If key exists but has no expiration → (-1, nil)
//   - Otherwise → (remaining TTL, nil)
func NormalizeTTLResult(duration time.Duration, keyExists bool) time.Duration {
	if !keyExists {
		// Key does not exist
		return 0
	}

	switch duration {
	case -1 * time.Second:
		// Exists but no TTL (Redis convention)
		return -1
	case -2 * time.Second:
		// Does not exist (shouldn’t reach here if keyExists==true)
		return 0
	default:
		if duration < 0 {
			// Unknown negative duration, treat as no expiration
			return -1
		}
		return duration
	}
}

// NormalizeExistsResult converts backend-specific existence results
// into a unified count value.
//
// Rules:
//   - Redis returns the number of keys found (e.g. 0, 1, 2, ...).
//   - Memcache can only check key-by-key, so it's usually 0 or 1.
//   - Any negative or invalid input will return 0 (not found).
//
// Unified Output:
//   - int64: number of existing keys
func NormalizeExistsResult(count int64) int64 {
	if count < 0 {
		return 0
	}
	return count
}

// NormalizeIncrementResult ensures consistent increment/decrement behavior across cache drivers.
//
// Parameters:
//   - value: numeric value returned from backend
//   - err: backend error, if any
//   - isCacheMiss: indicates whether the key was missing
//   - startIfMissing: starting value to use if key didn't exist
//
// Returns:
//   - normalized new value
//   - normalized error (nil if key was missing but normalized)
func NormalizeIncrementResult(value int64, err error, isCacheMiss bool, startIfMissing int64) (int64, error) {
	if err == nil {
		return value, nil
	}

	if isCacheMiss {
		// Start fresh if the key doesn’t exist
		return startIfMissing, nil
	}

	// Pass through any other error (e.g., type mismatch, connection failure)
	return 0, err
}

var (
	ErrCacheMiss = errors.New("cache miss")
)

// NormalizeError translates driver-specific cache-miss errors into a unified one.
func NormalizeError(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "memcache: cache miss") ||
		strings.Contains(err.Error(), "redis: nil") {
		return ErrCacheMiss
	}

	return err
}

// IsCacheMiss is a convenience checker
func IsCacheMiss(err error) bool {
	return errors.Is(NormalizeError(err), ErrCacheMiss)
}
