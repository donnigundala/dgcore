package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/donnigundala/dg-core/errors"
)

// Timeout returns a middleware that times out requests after the specified duration.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Run the handler in a goroutine
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				err := errors.New("request timeout").
					WithCode("REQUEST_TIMEOUT").
					WithStatus(http.StatusRequestTimeout).
					WithField("timeout", timeout.String())
				errors.WriteHTTPError(w, err)
			}
		})
	}
}
