package logging

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/donnigundala/dg-core/ctxutil"
)

// Logger wraps slog.Logger with additional functionality.
type Logger struct {
	logger *slog.Logger
	config Config
}

// Config defines the logger configuration.
type Config struct {
	Level      slog.Level
	Output     io.Writer
	JSONFormat bool
	AddSource  bool
}

// DefaultConfig returns the default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:      slog.LevelInfo,
		Output:     os.Stdout,
		JSONFormat: false,
		AddSource:  false,
	}
}

// New creates a new Logger with the given configuration.
func New(config Config) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	if config.JSONFormat {
		handler = slog.NewJSONHandler(config.Output, opts)
	} else {
		handler = slog.NewTextHandler(config.Output, opts)
	}

	return &Logger{
		logger: slog.New(handler),
		config: config,
	}
}

// NewDefault creates a new Logger with default configuration.
func NewDefault() *Logger {
	return New(DefaultConfig())
}

// WithContext returns a logger that extracts context information (request ID, user, etc.).
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger

	// Add request ID if available
	if requestID := ctxutil.RequestIDFromContext(ctx); requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	// You can add more context fields here
	// e.g., user ID, trace ID, etc.

	return &Logger{
		logger: logger,
		config: l.config,
	}
}

// WithFields returns a logger with additional fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		logger: l.logger.With(args...),
		config: l.config,
	}
}

// With returns a logger with additional key-value pairs.
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
		config: l.config,
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// DebugContext logs a debug message with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithContext(ctx).Debug(msg, args...)
}

// InfoContext logs an info message with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithContext(ctx).Info(msg, args...)
}

// WarnContext logs a warning message with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithContext(ctx).Warn(msg, args...)
}

// ErrorContext logs an error message with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithContext(ctx).Error(msg, args...)
}

// Underlying returns the underlying slog.Logger.
func (l *Logger) Underlying() *slog.Logger {
	return l.logger
}

// Global logger instance
var defaultLogger = NewDefault()

// SetDefault sets the default global logger.
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// Default returns the default global logger.
func Default() *Logger {
	return defaultLogger
}

// Global convenience functions

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message using the default logger.
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message using the default logger.
func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

// WithContext returns a logger with context using the default logger.
func WithContext(ctx context.Context) *Logger {
	return defaultLogger.WithContext(ctx)
}

// WithFields returns a logger with fields using the default logger.
func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}
