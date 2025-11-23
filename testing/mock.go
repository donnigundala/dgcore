package testing

import (
	"context"
	"log/slog"
	"sync"

	"github.com/donnigundala/dg-core/logging"
)

// MockContainer is a simple mock container for testing.
type MockContainer struct {
	bindings map[string]interface{}
	mu       sync.RWMutex
}

// NewMockContainer creates a new mock container.
func NewMockContainer() *MockContainer {
	return &MockContainer{
		bindings: make(map[string]interface{}),
	}
}

// Set sets a binding in the mock container.
func (m *MockContainer) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bindings[key] = value
}

// Get retrieves a binding from the mock container.
func (m *MockContainer) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.bindings[key]
	return value, ok
}

// MockLogger is a mock logger for testing.
type MockLogger struct {
	logs []LogEntry
	mu   sync.Mutex
}

// LogEntry represents a log entry.
type LogEntry struct {
	Level   string
	Message string
	Args    []interface{}
}

// NewMockLogger creates a new mock logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs: make([]LogEntry, 0),
	}
}

// Info logs an info message.
func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.log("info", msg, args...)
}

// Debug logs a debug message.
func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.log("debug", msg, args...)
}

// Warn logs a warning message.
func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.log("warn", msg, args...)
}

// Error logs an error message.
func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.log("error", msg, args...)
}

// InfoContext logs an info message with context.
func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	m.log("info", msg, args...)
}

// DebugContext logs a debug message with context.
func (m *MockLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	m.log("debug", msg, args...)
}

// WarnContext logs a warning message with context.
func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	m.log("warn", msg, args...)
}

// ErrorContext logs an error message with context.
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	m.log("error", msg, args...)
}

// WithFields returns a logger with fields.
func (m *MockLogger) WithFields(fields map[string]interface{}) *logging.Logger {
	// Return a new logger for compatibility
	return logging.New(logging.Config{
		Level:  slog.LevelInfo,
		Output: nil,
	})
}

// WithContext returns a logger with context.
func (m *MockLogger) WithContext(ctx context.Context) *logging.Logger {
	// Return a new logger for compatibility
	return logging.New(logging.Config{
		Level:  slog.LevelInfo,
		Output: nil,
	})
}

func (m *MockLogger) log(level, msg string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, LogEntry{
		Level:   level,
		Message: msg,
		Args:    args,
	})
}

// AssertLogged asserts that a message was logged at a specific level.
func (m *MockLogger) AssertLogged(level, message string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.logs {
		if entry.Level == level && entry.Message == message {
			return true
		}
	}
	return false
}

// AssertNotLogged asserts that a message was not logged at a specific level.
func (m *MockLogger) AssertNotLogged(level string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.logs {
		if entry.Level == level {
			return false
		}
	}
	return true
}

// GetLogs returns all logged entries.
func (m *MockLogger) GetLogs() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]LogEntry{}, m.logs...)
}

// Clear clears all logged entries.
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = make([]LogEntry, 0)
}

// MockConfig is a mock configuration for testing.
type MockConfig struct {
	values map[string]interface{}
	mu     sync.RWMutex
}

// NewMockConfig creates a new mock config.
func NewMockConfig(values map[string]interface{}) *MockConfig {
	return &MockConfig{
		values: values,
	}
}

// Get retrieves a config value.
func (m *MockConfig) Get(key string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values[key]
}

// Set sets a config value.
func (m *MockConfig) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = value
}
