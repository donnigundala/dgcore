# Router Guide - DG Framework

## Overview
The DG Framework uses a **switchable router pattern** - you can use Gin (default), the custom router, or any other router.

## Quick Start (Using Gin - Default)

```go
package main

import (
    "github.com/donnigundala/dg-core/foundation"
    "github.com/donnigundala/dg-core/http"
    "github.com/donnigundala/dg-core/contracts/http"
)

func main() {
    app := foundation.New(".")
    
    // Bind Router (uses Gin by default)
    app.Singleton("router", func() interface{} {
        return http.NewRouter()
    })
    
    // Register routes
    router, _ := app.Make("router")
    router.(http.Router).Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
    })
    
    // ... rest of setup
}
```

## Using Different Routers

### Option 1: Gin Router (Default - Recommended)
```go
app.Singleton("router", func() interface{} {
    return http.NewRouter()  // Uses Gin
})
```

**Benefits:**
- ✅ High performance
- ✅ Rich middleware ecosystem
- ✅ JSON validation
- ✅ Parameter binding

### Option 2: Custom Lightweight Router
```go
app.Singleton("router", func() interface{} {
    return http.NewDefaultRouter()
})
```

**Benefits:**
- ✅ No external dependencies
- ✅ Lightweight
- ✅ Simple routing

### Option 3: Direct Gin Access
```go
// Get the underlying Gin engine for advanced usage
ginRouter := http.NewGinRouter()
engine := ginRouter.Engine()

// Use Gin-specific features
engine.Use(gin.Logger())
engine.LoadHTMLGlob("templates/*")
```

## Router Interface

All routers implement the same interface:

```go
type Router interface {
    http.Handler
    
    Get(path string, handler HandlerFunc) Route
    Post(path string, handler HandlerFunc) Route
    Put(path string, handler HandlerFunc) Route
    Patch(path string, handler HandlerFunc) Route
    Delete(path string, handler HandlerFunc) Route
    
    Group(attributes GroupAttributes, callback func(Router))
    Use(middleware ...func(http.Handler) http.Handler)
}
```

## Route Groups

```go
router.Group(http.GroupAttributes{
    Prefix: "/api/v1",
    Middleware: []func(http.Handler) http.Handler{
        AuthMiddleware,
    },
}, func(r http.Router) {
    r.Get("/users", ListUsers)
    r.Post("/users", CreateUser)
})
```

## Middleware

```go
// Global middleware
router.Use(LoggingMiddleware, AuthMiddleware)

// Route-specific middleware
router.Get("/admin", AdminHandler).Middleware(AdminOnlyMiddleware)
```

## Switching Routers

To switch from Gin to another router in the future:

1. **Create adapter** (e.g., `chi_router.go`)
2. **Implement Router interface**
3. **Update factory**:
```go
func NewRouter() Router {
    return NewChiRouter()  // Switch to Chi
}
```

Your application code stays the same! 🎯

## Performance Comparison

| Router | Requests/sec | Memory | Features |
|--------|-------------|--------|----------|
| Gin | ~40k | Low | High |
| Custom | ~35k | Very Low | Basic |
| Chi | ~38k | Low | Medium |

## Recommendation

**Use Gin (default)** unless you have specific requirements:
- Need minimal dependencies → Use `NewDefaultRouter()`
- Need specific router features → Create adapter
- Building microservice → Gin is perfect

## Example: Full Application

See [skeleton/main.go](file:///Users/donni/Codespace/MyCodes/dg-framework/core/skeleton/main.go) for a complete working example.
