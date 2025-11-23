package http

import "net/http"

// Kernel defines the interface for the HTTP kernel.
// It is responsible for handling incoming HTTP requests and managing the global middleware stack.
type Kernel interface {
	http.Handler

	// Bootstrap bootstraps the application for HTTP requests.
	Bootstrap()

	// GetMiddleware returns the global middleware stack.
	GetMiddleware() []func(http.Handler) http.Handler
}
