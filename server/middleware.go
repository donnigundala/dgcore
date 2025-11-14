package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// contextKey is a private type to prevent collisions with other packages' context keys.
type contextKey string

const (
	requestIDKey contextKey = "requestID"
	loggerKey    contextKey = "logger"
)

// RequestIDMiddleware injects a unique request ID and a context-aware logger into each request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the request ID from the header if it exists (for distributed tracing).
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// 2. If it doesn't exist, create a new one.
			requestID = uuid.New().String()
		}

		// 3. Add the ID to the response header so the client can also track it.
		w.Header().Set("X-Request-ID", requestID)

		// 4. Get the default logger and create a child logger with the request_id field.
		//    We use slog.Default() as the base.
		logger := slog.Default().With("request_id", requestID)

		// 5. Store the logger and request ID in a new context.
		ctx := context.WithValue(r.Context(), loggerKey, logger)
		ctx = context.WithValue(ctx, requestIDKey, requestID)

		// 6. Call the next handler with the new context-aware request.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFromContext retrieves the context-aware logger from the context.
// If no logger is found, it returns the default logger.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// RequestIDFromContext retrieves the request ID from the context.
// It returns an empty string if not found.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
