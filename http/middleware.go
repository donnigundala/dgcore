package http

import (
	"log/slog"
	"net/http"

	"github.com/donnigundala/dgcore/ctxutil"
)

// RequestIDMiddleware injects a unique request ID and a context-aware logger into each request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the request ID from the header or create a new one.
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = ctxutil.NewRequestID()
		}

		// Add the ID to the response header so the client can also track it.
		w.Header().Set("X-Request-ID", requestID)

		// Create a child logger with the request_id field.
		logger := slog.Default().With("request_id", requestID)

		// Store the new logger and the request ID in the context using helpers from ctxutil.
		ctx := r.Context()
		ctx = ctxutil.WithLogger(ctx, logger)
		ctx = ctxutil.WithRequestID(ctx, requestID)

		// Call the next handler with the new context-aware request.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
