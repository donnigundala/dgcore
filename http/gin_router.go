package http

import (
	"net/http"

	contractHTTP "github.com/donnigundala/dg-core/contracts/http"
	"github.com/gin-gonic/gin"
)

// GinRouter is a Gin-based implementation of the Router interface.
type GinRouter struct {
	engine *gin.Engine
}

// NewGinRouter creates a new Gin-based router.
func NewGinRouter() *GinRouter {
	// Create Gin engine without default middleware
	engine := gin.New()

	// Add recovery middleware (recommended)
	engine.Use(gin.Recovery())

	return &GinRouter{
		engine: engine,
	}
}

// Get registers a GET route.
func (g *GinRouter) Get(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.engine.GET(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "GET"}
}

// Post registers a POST route.
func (g *GinRouter) Post(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.engine.POST(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "POST"}
}

// Put registers a PUT route.
func (g *GinRouter) Put(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.engine.PUT(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "PUT"}
}

// Patch registers a PATCH route.
func (g *GinRouter) Patch(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.engine.PATCH(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "PATCH"}
}

// Delete registers a DELETE route.
func (g *GinRouter) Delete(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.engine.DELETE(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "DELETE"}
}

// Group creates a route group.
func (g *GinRouter) Group(attributes contractHTTP.GroupAttributes, callback func(contractHTTP.Router)) {
	// Create Gin group
	ginGroup := g.engine.Group(attributes.Prefix)

	// Apply group middleware
	for _, mw := range attributes.Middleware {
		ginGroup.Use(ginMiddlewareAdapter(mw))
	}

	// Create a wrapper router for the group
	groupRouter := &ginGroupRouter{
		group:  ginGroup,
		engine: g.engine,
	}

	// Execute callback with group router
	callback(groupRouter)
}

// Use adds global middleware to the router.
func (g *GinRouter) Use(middleware ...func(http.Handler) http.Handler) {
	for _, mw := range middleware {
		g.engine.Use(ginMiddlewareAdapter(mw))
	}
}

// ServeHTTP implements http.Handler interface.
func (g *GinRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.engine.ServeHTTP(w, r)
}

// Engine returns the underlying Gin engine for advanced usage.
func (g *GinRouter) Engine() *gin.Engine {
	return g.engine
}

// ginMiddlewareAdapter converts standard http middleware to Gin middleware.
func ginMiddlewareAdapter(mw func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

// ginGroupRouter wraps a Gin RouterGroup to implement the Router interface.
type ginGroupRouter struct {
	group  *gin.RouterGroup
	engine *gin.Engine
}

func (g *ginGroupRouter) Get(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.group.GET(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "GET"}
}

func (g *ginGroupRouter) Post(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.group.POST(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "POST"}
}

func (g *ginGroupRouter) Put(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.group.PUT(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "PUT"}
}

func (g *ginGroupRouter) Patch(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.group.PATCH(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "PATCH"}
}

func (g *ginGroupRouter) Delete(path string, handler contractHTTP.HandlerFunc) contractHTTP.Route {
	g.group.DELETE(path, gin.WrapF(handler))
	return &ginRoute{path: path, method: "DELETE"}
}

func (g *ginGroupRouter) Group(attributes contractHTTP.GroupAttributes, callback func(contractHTTP.Router)) {
	nestedGroup := g.group.Group(attributes.Prefix)

	for _, mw := range attributes.Middleware {
		nestedGroup.Use(ginMiddlewareAdapter(mw))
	}

	groupRouter := &ginGroupRouter{
		group:  nestedGroup,
		engine: g.engine,
	}

	callback(groupRouter)
}

func (g *ginGroupRouter) Use(middleware ...func(http.Handler) http.Handler) {
	for _, mw := range middleware {
		g.group.Use(ginMiddlewareAdapter(mw))
	}
}

func (g *ginGroupRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.engine.ServeHTTP(w, r)
}
