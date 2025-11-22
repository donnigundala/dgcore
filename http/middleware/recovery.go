package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/donnigundala/dgcore/errors"
	"github.com/donnigundala/dgcore/logging"
)

// Recovery returns a middleware that recovers from panics.
func Recovery(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					stack := string(debug.Stack())

					logger.ErrorContext(r.Context(), "Panic recovered",
						"error", err,
						"stack", stack,
						"method", r.Method,
						"path", r.URL.Path,
					)

					// Create error response
					wrappedErr := errors.New(fmt.Sprintf("internal server error: %v", err)).
						WithCode("PANIC_RECOVERED").
						WithStatus(http.StatusInternalServerError)

					// Write error response
					errors.WriteHTTPError(w, wrappedErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryWithDefault returns a Recovery middleware using the default logger.
func RecoveryWithDefault() func(http.Handler) http.Handler {
	return Recovery(logging.Default())
}
