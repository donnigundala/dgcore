package ctxutil

import (
	"context"
	"fmt"
	"strconv"
)

// GetString retrieves a string value from context with a default fallback.
func GetString(ctx context.Context, key string, defaultValue string) string {
	if val := ctx.Value(contextKey(key)); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetInt retrieves an int value from context with a default fallback.
func GetInt(ctx context.Context, key string, defaultValue int) int {
	if val := ctx.Value(contextKey(key)); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// GetInt64 retrieves an int64 value from context with a default fallback.
func GetInt64(ctx context.Context, key string, defaultValue int64) int64 {
	if val := ctx.Value(contextKey(key)); val != nil {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// GetBool retrieves a bool value from context with a default fallback.
func GetBool(ctx context.Context, key string, defaultValue bool) bool {
	if val := ctx.Value(contextKey(key)); val != nil {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		}
	}
	return defaultValue
}

// GetFloat64 retrieves a float64 value from context with a default fallback.
func GetFloat64(ctx context.Context, key string, defaultValue float64) float64 {
	if val := ctx.Value(contextKey(key)); val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

// Set stores a value in the context with the given key.
func Set(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, contextKey(key), value)
}

// Get retrieves a value from context.
func Get(ctx context.Context, key string) interface{} {
	return ctx.Value(contextKey(key))
}

// MustGetString retrieves a string value from context or panics if not found.
func MustGetString(ctx context.Context, key string) string {
	if val := ctx.Value(contextKey(key)); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	panic(fmt.Sprintf("context key %s not found or not a string", key))
}

// MustGetInt retrieves an int value from context or panics if not found.
func MustGetInt(ctx context.Context, key string) int {
	if val := ctx.Value(contextKey(key)); val != nil {
		if i, ok := val.(int); ok {
			return i
		}
	}
	panic(fmt.Sprintf("context key %s not found or not an int", key))
}
