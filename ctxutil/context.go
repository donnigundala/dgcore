package ctxutil

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// contextKey is a private type to prevent collisions with other packages' context keys.
type contextKey string

const (
	requestIDKey contextKey = "requestID"
	loggerKey    contextKey = "logger"
)

// WithLogger returns a new context that holds the given logger.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext retrieves the logger from the context.
// If no logger is found, it returns the default logger.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithRequestID returns a new context that holds the given request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext retrieves the request ID from the context.
// It returns an empty string if not found.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// NewRequestID generates a new unique request ID.
func NewRequestID() string {
	return uuid.New().String()
}
