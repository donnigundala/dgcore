package http

import contractHTTP "github.com/donnigundala/dgcore/contracts/http"

// NewRouter creates a new router instance.
// By default, this returns a Gin-based router for better performance and features.
func NewRouter() contractHTTP.Router {
	return NewGinRouter()
}

// NewDefaultRouter creates the framework's custom lightweight router.
// Use this if you want a minimal router without external dependencies.
func NewDefaultRouter() contractHTTP.Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}
