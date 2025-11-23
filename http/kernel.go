package http

import (
	"net/http"

	"github.com/donnigundala/dgcore/contracts/foundation"
	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
)

// Kernel is the concrete implementation of the HTTP kernel.
type Kernel struct {
	app        foundation.Application
	router     contractHTTP.Router
	middleware []func(http.Handler) http.Handler
}

// NewKernel creates a new Kernel instance.
func NewKernel(app foundation.Application, router contractHTTP.Router) *Kernel {
	return &Kernel{
		app:        app,
		router:     router,
		middleware: []func(http.Handler) http.Handler{
			// Default global middleware can be added here
		},
	}
}

// Bootstrap bootstraps the application for HTTP requests.
func (k *Kernel) Bootstrap() {
	if !k.app.IsBooted() {
		k.app.Boot()
	}
}

// GetMiddleware returns the global middleware stack.
func (k *Kernel) GetMiddleware() []func(http.Handler) http.Handler {
	return k.middleware
}

// Handle handles the incoming HTTP request.
func (k *Kernel) Handle(w http.ResponseWriter, r *http.Request) {
	// 1. Bootstrap the application
	k.Bootstrap()

	// 2. Construct Handler Chain
	// Global Middleware -> Router

	var handler http.Handler = k.router

	// Apply Global Middleware (in reverse order)
	for i := len(k.middleware) - 1; i >= 0; i-- {
		handler = k.middleware[i](handler)
	}

	// 3. Serve
	handler.ServeHTTP(w, r)
}

// ServeHTTP satisfies the http.Handler interface.
func (k *Kernel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k.Handle(w, r)
}
