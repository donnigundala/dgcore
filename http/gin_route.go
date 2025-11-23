package http

import (
	"net/http"

	contractHTTP "github.com/donnigundala/dg-core/contracts/http"
)

// ginRoute implements the Route interface for Gin routes.
type ginRoute struct {
	path   string
	method string
}

// Name sets the route name (not used in Gin, but satisfies interface).
func (r *ginRoute) Name(name string) contractHTTP.Route {
	// Gin doesn't support named routes directly
	// This is a no-op for compatibility
	return r
}

// Middleware adds middleware to the route (not used in Gin, but satisfies interface).
func (r *ginRoute) Middleware(middleware ...func(http.Handler) http.Handler) contractHTTP.Route {
	// Middleware is applied at registration time in Gin
	// This is a no-op for compatibility
	return r
}
