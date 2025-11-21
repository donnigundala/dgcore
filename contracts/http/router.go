package http

import "net/http"

// HandlerFunc is a function that handles an HTTP request.
// It matches the standard http.HandlerFunc.
type HandlerFunc = http.HandlerFunc

// GroupAttributes defines the attributes for a route group.
type GroupAttributes struct {
	Prefix     string
	Middleware []func(http.Handler) http.Handler
}

// Router defines the interface for the HTTP router.
type Router interface {
	http.Handler

	// HTTP Verbs
	Get(path string, handler HandlerFunc) Route
	Post(path string, handler HandlerFunc) Route
	Put(path string, handler HandlerFunc) Route
	Patch(path string, handler HandlerFunc) Route
	Delete(path string, handler HandlerFunc) Route

	// Grouping
	Group(attributes GroupAttributes, callback func(Router))

	// Middleware
	Use(middleware ...func(http.Handler) http.Handler)
}

// Route defines the interface for a registered route.
// This allows for fluent chaining of methods like Name(), Middleware(), etc.
type Route interface {
	Name(name string) Route
	Middleware(middleware ...func(http.Handler) http.Handler) Route
}
