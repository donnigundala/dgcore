# Middleware Guide - DG Framework

## Overview
The DG Framework includes **7 essential production middleware** components to handle common concerns like CORS, security, rate limiting, and more.

## Available Middleware

### 1. Recovery 🛡️ (Critical)
**Purpose:** Catch panics and return proper error responses

**Features:**
- Catches panics in handlers
- Logs stack trace with context
- Returns 500 error response
- Integrates with logging package

**Usage:**
```go
import coreHTTP "github.com/donnigundala/dg-core/http"

// With custom logger
router.Use(coreHTTP.Recovery(logger))

// With default logger
router.Use(coreHTTP.RecoveryWithDefault())
```

---

### 2. CORS 🌐 (Critical)
**Purpose:** Handle Cross-Origin Resource Sharing for APIs

**Features:**
- Configurable allowed origins
- Allowed methods and headers
- Credentials support
- Preflight request handling
- Wildcard subdomain support (*.example.com)

**Usage:**
```go
// Default config (allows all origins)
router.Use(coreHTTP.CORSWithDefault())

// Custom config
router.Use(coreHTTP.CORS(coreHTTP.CORSConfig{
    AllowedOrigins: []string{"https://example.com", "*.myapp.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge: 3600,
}))
```

---

### 3. Security Headers 🔒 (Critical)
**Purpose:** Add security headers to protect against common attacks

**Features:**
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security (HSTS)
- Content-Security-Policy (CSP)
- Referrer-Policy

**Usage:**
```go
// Default config
router.Use(coreHTTP.SecurityHeadersWithDefault())

// Custom config
router.Use(coreHTTP.SecurityHeaders(coreHTTP.SecurityConfig{
    ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'",
    XFrameOptions: "SAMEORIGIN",
    HSTSMaxAge: 31536000,
    HSTSIncludeSubdomains: true,
}))
```

---

### 4. Rate Limiting ⏱️ (Important)
**Purpose:** Prevent abuse by limiting requests per client

**Features:**
- Per-IP rate limiting
- Configurable requests per second
- Burst size support
- Custom key function (e.g., by user ID)
- Handles X-Forwarded-For and X-Real-IP

**Usage:**
```go
// Default config (10 req/s, burst 20)
router.Use(coreHTTP.RateLimitWithDefault())

// Custom config
router.Use(coreHTTP.RateLimit(coreHTTP.RateLimitConfig{
    RequestsPerSecond: 5,
    BurstSize: 10,
    KeyFunc: func(r *http.Request) string {
        // Custom key extraction (e.g., by user ID)
        return r.Header.Get("X-User-ID")
    },
}))
```

---

### 5. Request Timeout ⏰ (Important)
**Purpose:** Prevent hanging requests

**Features:**
- Configurable timeout duration
- Context cancellation
- Returns 408 Request Timeout

**Usage:**
```go
import "time"

// 30 second timeout
router.Use(coreHTTP.Timeout(30 * time.Second))
```

---

### 6. Body Size Limit 📏 (Important)
**Purpose:** Prevent large payload attacks

**Features:**
- Configurable max body size
- Returns 413 Payload Too Large
- Applies to POST/PUT/PATCH only
- Checks Content-Length header

**Usage:**
```go
// 10MB limit
router.Use(coreHTTP.BodySizeLimit(10 * 1024 * 1024))

// With error response
router.Use(coreHTTP.BodySizeLimitWithError(10 * 1024 * 1024))
```

---

### 7. Compression 🗜️ (Optional)
**Purpose:** Compress responses to reduce bandwidth

**Features:**
- Gzip compression
- Configurable compression level
- Content-type filtering
- Minimum size threshold
- Only compresses if client accepts gzip

**Usage:**
```go
// Default config
router.Use(coreHTTP.CompressWithDefault())

// Custom config
router.Use(coreHTTP.Compress(coreHTTP.CompressConfig{
    Level: gzip.BestCompression,
    MinSize: 2048, // 2KB
    ContentTypes: []string{
        "text/",
        "application/json",
        "application/javascript",
    },
}))
```

---

## Complete Example

```go
package main

import (
    "time"
    
    "github.com/donnigundala/dg-core/foundation"
    coreHTTP "github.com/donnigundala/dg-core/http"
    "github.com/donnigundala/dg-core/logging"
)

func main() {
    app := foundation.New(".")
    logger := logging.NewDefault()
    
    // Bind router
    app.Singleton("router", func() interface{} {
        return coreHTTP.NewRouter()
    })
    
    router, _ := app.Make("router")
    r := router.(coreHTTP.Router)
    
    // Apply global middleware (order matters!)
    r.Use(
        coreHTTP.Recovery(logger),              // 1. Catch panics
        coreHTTP.CORSWithDefault(),             // 2. CORS
        coreHTTP.SecurityHeadersWithDefault(),  // 3. Security headers
        coreHTTP.BodySizeLimit(10 * 1024 * 1024), // 4. Body size limit
        coreHTTP.CompressWithDefault(),         // 5. Compression (last)
    )
    
    // Apply route-specific middleware
    r.Group(coreHTTP.GroupAttributes{
        Prefix: "/api",
        Middleware: []func(http.Handler) http.Handler{
            coreHTTP.RateLimit(coreHTTP.RateLimitConfig{
                RequestsPerSecond: 10,
                BurstSize: 20,
            }),
            coreHTTP.Timeout(30 * time.Second),
        },
    }, func(r coreHTTP.Router) {
        r.Get("/users", ListUsers)
        r.Post("/users", CreateUser)
    })
    
    // Start server...
}
```

## Middleware Order

**Recommended order:**
1. **Recovery** - Catch panics first
2. **CORS** - Handle preflight requests early
3. **Security Headers** - Add security headers
4. **Rate Limiting** - Prevent abuse
5. **Timeout** - Set request timeout
6. **Body Size Limit** - Check payload size
7. **Compression** - Compress response (last)

## Best Practices

### Production Setup
```go
// Strict CORS
router.Use(coreHTTP.CORS(coreHTTP.CORSConfig{
    AllowedOrigins: []string{"https://myapp.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowCredentials: true,
}))

// Strong security headers
router.Use(coreHTTP.SecurityHeaders(coreHTTP.SecurityConfig{
    ContentSecurityPolicy: "default-src 'self'",
    HSTSMaxAge: 31536000,
    HSTSIncludeSubdomains: true,
}))

// Conservative rate limiting
router.Use(coreHTTP.RateLimit(coreHTTP.RateLimitConfig{
    RequestsPerSecond: 5,
    BurstSize: 10,
}))
```

### Development Setup
```go
// Permissive CORS
router.Use(coreHTTP.CORSWithDefault())

// Relaxed security
router.Use(coreHTTP.SecurityHeadersWithDefault())

// Higher rate limits
router.Use(coreHTTP.RateLimit(coreHTTP.RateLimitConfig{
    RequestsPerSecond: 100,
    BurstSize: 200,
}))
```

## Testing Middleware

```bash
# Test CORS
curl -H "Origin: https://example.com" \
     -H "Access-Control-Request-Method: POST" \
     -X OPTIONS http://localhost:8080/api/users

# Test Rate Limiting
for i in {1..30}; do curl http://localhost:8080/api/users; done

# Test Compression
curl -H "Accept-Encoding: gzip" http://localhost:8080/api/users

# Test Body Size Limit
curl -X POST -d @large-file.json http://localhost:8080/api/users
```

## Summary

✅ **7 Essential Middleware** - Production-ready  
✅ **Easy Configuration** - Sensible defaults  
✅ **Composable** - Mix and match  
✅ **Type-Safe** - Full Go type safety  
✅ **Well-Tested** - Reliable implementations
