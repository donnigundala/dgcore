package ctxutil

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

// TestWithLogger_Storage tests storing logger in context
func TestWithLogger_Storage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx = WithLogger(ctx, logger)

	retrieved := LoggerFromContext(ctx)
	if retrieved != logger {
		t.Error("expected to retrieve the same logger instance")
	}
}

// TestLoggerFromContext_Retrieval tests retrieving logger from context
func TestLoggerFromContext_Retrieval(t *testing.T) {
	// Test with logger in context
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx = WithLogger(ctx, logger)

	retrieved := LoggerFromContext(ctx)
	if retrieved != logger {
		t.Error("expected to retrieve the same logger")
	}

	// Test with no logger in context (should return default)
	emptyCtx := context.Background()
	defaultLogger := LoggerFromContext(emptyCtx)

	if defaultLogger == nil {
		t.Error("expected default logger, got nil")
	}

	if defaultLogger != slog.Default() {
		t.Error("expected default logger when none in context")
	}
}

// TestWithRequestID_Storage tests storing request ID in context
func TestWithRequestID_Storage(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-id-123"

	ctx = WithRequestID(ctx, requestID)

	retrieved := RequestIDFromContext(ctx)
	if retrieved != requestID {
		t.Errorf("expected request ID %q, got %q", requestID, retrieved)
	}
}

// TestRequestIDFromContext_NotFound tests retrieving non-existent request ID
func TestRequestIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	retrieved := RequestIDFromContext(ctx)
	if retrieved != "" {
		t.Errorf("expected empty string, got %q", retrieved)
	}
}

// TestNewRequestID_Unique tests that generated request IDs are unique
func TestNewRequestID_Unique(t *testing.T) {
	id1 := NewRequestID()
	id2 := NewRequestID()

	if id1 == "" {
		t.Error("expected non-empty request ID")
	}

	if id2 == "" {
		t.Error("expected non-empty request ID")
	}

	if id1 == id2 {
		t.Error("expected unique request IDs")
	}

	// Check format (should be UUID)
	if len(id1) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars", len(id1))
	}
}

// TestGetString_WithValue tests retrieving string from context
func TestGetString_WithValue(t *testing.T) {
	ctx := context.Background()
	ctx = Set(ctx, "key", "value")

	result := GetString(ctx, "key", "default")
	if result != "value" {
		t.Errorf("expected %q, got %q", "value", result)
	}
}

// TestGetString_WithDefault tests default value when key not found
func TestGetString_WithDefault(t *testing.T) {
	ctx := context.Background()

	result := GetString(ctx, "nonexistent", "default")
	if result != "default" {
		t.Errorf("expected %q, got %q", "default", result)
	}
}

// TestGetInt_WithValue tests retrieving int from context
func TestGetInt_WithValue(t *testing.T) {
	ctx := context.Background()
	ctx = Set(ctx, "count", 42)

	result := GetInt(ctx, "count", 0)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

// TestGetInt_TypeConversion tests int type conversions
func TestGetInt_TypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected int
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"string", "123", 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Set(context.Background(), "key", tt.value)
			result := GetInt(ctx, "key", 0)

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestGetInt_WithDefault tests default value for int
func TestGetInt_WithDefault(t *testing.T) {
	ctx := context.Background()

	result := GetInt(ctx, "nonexistent", 999)
	if result != 999 {
		t.Errorf("expected 999, got %d", result)
	}
}

// TestGetInt64_WithValue tests retrieving int64 from context
func TestGetInt64_WithValue(t *testing.T) {
	ctx := context.Background()
	ctx = Set(ctx, "bignum", int64(9223372036854775807))

	result := GetInt64(ctx, "bignum", 0)
	if result != 9223372036854775807 {
		t.Errorf("expected 9223372036854775807, got %d", result)
	}
}

// TestGetInt64_TypeConversion tests int64 type conversions
func TestGetInt64_TypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected int64
	}{
		{"int64", int64(100), 100},
		{"int", 42, 42},
		{"string", "456", 456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Set(context.Background(), "key", tt.value)
			result := GetInt64(ctx, "key", 0)

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestGetBool_WithValue tests retrieving bool from context
func TestGetBool_WithValue(t *testing.T) {
	ctx := context.Background()
	ctx = Set(ctx, "enabled", true)

	result := GetBool(ctx, "enabled", false)
	if result != true {
		t.Error("expected true, got false")
	}
}

// TestGetBool_TypeConversion tests bool type conversions
func TestGetBool_TypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"bool_true", true, true},
		{"bool_false", false, false},
		{"string_true", "true", true},
		{"string_false", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Set(context.Background(), "key", tt.value)
			result := GetBool(ctx, "key", false)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetFloat64_WithValue tests retrieving float64 from context
func TestGetFloat64_WithValue(t *testing.T) {
	ctx := context.Background()
	ctx = Set(ctx, "price", 99.99)

	result := GetFloat64(ctx, "price", 0.0)
	if result != 99.99 {
		t.Errorf("expected 99.99, got %f", result)
	}
}

// TestGetFloat64_TypeConversion tests float64 type conversions
func TestGetFloat64_TypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected float64
	}{
		{"float64", 3.14, 3.14},
		{"float32", float32(2.5), 2.5},
		{"string", "1.23", 1.23},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Set(context.Background(), "key", tt.value)
			result := GetFloat64(ctx, "key", 0.0)

			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// TestSet_And_Get tests basic set and get operations
func TestSet_And_Get(t *testing.T) {
	ctx := context.Background()

	// Test with various types
	ctx = Set(ctx, "string", "hello")
	ctx = Set(ctx, "int", 42)
	ctx = Set(ctx, "bool", true)

	if Get(ctx, "string") != "hello" {
		t.Error("expected to get string value")
	}

	if Get(ctx, "int") != 42 {
		t.Error("expected to get int value")
	}

	if Get(ctx, "bool") != true {
		t.Error("expected to get bool value")
	}

	// Test non-existent key
	if Get(ctx, "nonexistent") != nil {
		t.Error("expected nil for non-existent key")
	}
}

// TestMustGetString_Success tests successful retrieval
func TestMustGetString_Success(t *testing.T) {
	ctx := Set(context.Background(), "key", "value")

	result := MustGetString(ctx, "key")
	if result != "value" {
		t.Errorf("expected %q, got %q", "value", result)
	}
}

// TestMustGetString_Panic tests panic on missing key
func TestMustGetString_Panic(t *testing.T) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	MustGetString(ctx, "nonexistent")
}

// TestMustGetString_PanicWrongType tests panic on wrong type
func TestMustGetString_PanicWrongType(t *testing.T) {
	ctx := Set(context.Background(), "key", 123)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for wrong type, got none")
		}
	}()

	MustGetString(ctx, "key")
}

// TestMustGetInt_Success tests successful int retrieval
func TestMustGetInt_Success(t *testing.T) {
	ctx := Set(context.Background(), "key", 42)

	result := MustGetInt(ctx, "key")
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

// TestMustGetInt_Panic tests panic on missing key
func TestMustGetInt_Panic(t *testing.T) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	MustGetInt(ctx, "nonexistent")
}
