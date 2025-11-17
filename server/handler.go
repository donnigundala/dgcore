package server

import (
	"net/http"

	"github.com/donnigundala/dgcore/ctxutil"
)

// HandlerFunc is a custom handler function type that handles HTTP requests and can return an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP makes HandlerFunc implement the http.Handler interface.
// This allows for centralized error handling.
func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		// Retrieve the context-aware logger for consistent error logging.
		logger := ctxutil.LoggerFromContext(r.Context())
		logger.Error("unhandled error in handler", "error", err)

		// Send a generic HTTP 500 Internal Server Error response.
		// Avoid leaking error details to the client in a production environment.
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
