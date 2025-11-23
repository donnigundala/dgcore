package http

import (
	"net/http"
	"regexp"
	"strings"

	contractHTTP "github.com/donnigundala/dg-core/contracts/http"
)

// Router is the concrete implementation of the Router interface.
type Router struct {
	routes     []*Route
	middleware []func(http.Handler) http.Handler
	groups     []contractHTTP.GroupAttributes
}

// Get registers a GET route.
func (r *Router) Get(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	return r.addRoute(http.MethodGet, path, handler)
}

// Post registers a POST route.
func (r *Router) Post(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	return r.addRoute(http.MethodPost, path, handler)
}

// Put registers a PUT route.
func (r *Router) Put(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	return r.addRoute(http.MethodPut, path, handler)
}

// Patch registers a PATCH route.
func (r *Router) Patch(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	return r.addRoute(http.MethodPatch, path, handler)
}

// Delete registers a DELETE route.
func (r *Router) Delete(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	return r.addRoute(http.MethodDelete, path, handler)
}

// Group creates a route group.
func (r *Router) Group(attributes contractHTTP.GroupAttributes, callback func(contractHTTP.Router)) {
	// Push group attributes
	r.groups = append(r.groups, attributes)

	// Execute callback
	callback(r)

	// Pop group attributes
	r.groups = r.groups[:len(r.groups)-1]
}

// Use adds global middleware to the router.
func (r *Router) Use(middleware ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

// ServeHTTP handles the HTTP request.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 1. Match Route
	route, params := r.match(req.Method, req.URL.Path)
	if route == nil {
		http.NotFound(w, req)
		return
	}

	// TODO: Inject params into context
	_ = params

	// 2. Construct Handler Chain
	// Global Middleware -> Group Middleware -> Route Middleware -> Handler

	// Start with the final handler
	var handler http.Handler = route.handler

	// Apply Route Middleware (in reverse order)
	for i := len(route.middleware) - 1; i >= 0; i-- {
		handler = route.middleware[i](handler)
	}

	// Apply Global Middleware (in reverse order)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	// 3. Serve
	// TODO: Inject params into context
	handler.ServeHTTP(w, req)
}

// addRoute adds a route to the router, considering current groups.
func (r *Router) addRoute(method, path string, handler contractHTTP.HandlerFunc) *Route {
	fullPath := path
	var groupMiddleware []func(http.Handler) http.Handler

	// Apply groups
	for _, group := range r.groups {
		if group.Prefix != "" {
			fullPath = strings.TrimRight(group.Prefix, "/") + "/" + strings.TrimLeft(fullPath, "/")
		}
		groupMiddleware = append(groupMiddleware, group.Middleware...)
	}

	route := &Route{
		method:     method,
		path:       fullPath,
		handler:    handler,
		middleware: groupMiddleware,
	}

	r.routes = append(r.routes, route)
	return route
}

// match finds a matching route for the given method and path.
func (r *Router) match(method, path string) (*Route, map[string]string) {
	for _, route := range r.routes {
		if route.method != method {
			continue
		}

		// Simple exact match for now
		// TODO: Implement regex/parameter matching
		if route.path == path {
			return route, nil
		}

		// Basic parameter matching support (e.g. /users/{id})
		// This is a very naive implementation for demonstration.
		// Real implementation should use a more robust matcher.
		routePattern := route.path
		// Escape special regex characters in the path, but not { }
		// Actually, let's just replace {param} with ([^/]+)

		// Check if route has parameters
		if strings.Contains(routePattern, "{") {
			// Convert route path to regex
			// /users/{id} -> ^/users/([^/]+)$

			regexPattern := "^" + routePattern + "$"
			// Replace {param} with named capture group or just capture group
			// For simplicity, let's just use ([^/]+)

			// We need to handle multiple parameters
			re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
			matches := re.FindAllStringSubmatch(routePattern, -1)

			var paramNames []string
			for _, m := range matches {
				paramNames = append(paramNames, m[1])
			}

			regexPattern = re.ReplaceAllString(regexPattern, `([^/]+)`)

			matcher, err := regexp.Compile(regexPattern)
			if err != nil {
				continue
			}

			pathMatches := matcher.FindStringSubmatch(path)
			if pathMatches != nil {
				params := make(map[string]string)
				// pathMatches[0] is the full match
				// pathMatches[1:] are the submatches
				if len(pathMatches)-1 == len(paramNames) {
					for i, name := range paramNames {
						params[name] = pathMatches[i+1]
					}
					return route, params
				}
			}
		}
	}

	return nil, nil
}
