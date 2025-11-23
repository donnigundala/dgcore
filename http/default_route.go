package http

import (
	"net/http"

	contractHTTP "github.com/donnigundala/dg-core/contracts/http"
)

// Route represents a registered route in the router.
type Route struct {
	method     string
	path       string
	handler    contractHTTP.HandlerFunc
	name       string
	middleware []func(http.Handler) http.Handler
}

// Name sets the name of the route.
func (r *Route) Name(name string) contractHTTP.Route {
	r.name = name
	return r
}

// Middleware adds middleware to the route.
func (r *Route) Middleware(middleware ...func(http.Handler) http.Handler) contractHTTP.Route {
	r.middleware = append(r.middleware, middleware...)
	return r
}
