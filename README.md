# dg-core

> A pragmatic, modular Go framework for building production-ready applications with dependency injection, service providers, and a plugin system.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen)](coverage.out)

---

## Philosophy

dg-core follows **pragmatic minimalism** — a small, focused core with batteries included for common use cases. We provide essential building blocks (DI container, configuration, providers) plus commonly-needed features (HTTP, validation) to enable rapid development without sacrificing modularity.

**Read more:** [Framework Philosophy](docs/PHILOSOPHY.md)

---

## Features

### Core Foundation
- 🏗️ **Dependency Injection Container** - Type-safe, thread-safe DI with singleton and transient bindings
- ⚙️ **Configuration Management** - Environment-based config with YAML/ENV support
- 🔌 **Service Provider System** - Laravel-inspired provider pattern for modular architecture
- 🧩 **Plugin Architecture** - Extensible plugin system with metadata and dependency management
- ❌ **Error Handling** - Standardized errors with HTTP conversion and stack traces
- 📝 **Structured Logging** - Built on Go's `log/slog` with context integration

### HTTP & Web
- 🚀 **High-Performance Router** - Powered by Gin for speed and reliability
- 🛡️ **Production-Ready Middleware** - CORS, logging, recovery, security, rate limiting, compression
- ✅ **Request Validation** - Declarative validation with custom rules
- 🌐 **Context Utilities** - Request ID, client IP extraction, metadata helpers

### Developer Experience
- 🧪 **Testing Utilities** - Framework testing helpers and mocks
- 📚 **Comprehensive Documentation** - Guides, examples, and API reference
- 🎯 **Clear Error Messages** - Helpful errors for faster debugging
- 🔄 **Graceful Shutdown** - Proper resource cleanup and signal handling

---

## Quick Start

### Installation

```bash
go get github.com/donnigundala/dg-core
```

### Hello World

```go
package main

import (
    "net/http"
    "github.com/donnigundala/dg-core/foundation"
)

func main() {
    // Create application
    app := foundation.New(".")
    
    // Register HTTP routes
    router := app.Make("router").(*http.ServeMux)
    router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, dg-core!"))
    })
    
    // Boot and run
    app.Boot()
    http.ListenAndServe(":8080", router)
}
```

### With Service Providers

```go
package main

import (
    "github.com/donnigundala/dg-core/foundation"
    "yourapp/providers"
)

func main() {
    app := foundation.New(".")
    
    // Register service providers
    app.Register(providers.NewConfigProvider())
    app.Register(providers.NewDatabaseProvider())
    app.Register(providers.NewCacheProvider())
    
    // Boot application
    if err := app.Boot(); err != nil {
        panic(err)
    }
    
    // Start HTTP server
    // ...
}
```

---

## Core Concepts

### 1. Dependency Injection

```go
import "github.com/donnigundala/dg-core/container"

// Create container
c := container.NewContainer()

// Bind singleton
c.Singleton("db", func() interface{} {
    return &Database{Host: "localhost"}
})

// Bind transient
c.Bind("logger", func() interface{} {
    return NewLogger()
})

// Resolve
db, _ := c.Make("db")
```

### 2. Service Providers

```go
type CacheProvider struct {
    config CacheConfig
}

func (p *CacheProvider) Register(app foundation.Application) error {
    app.Singleton("cache", func() interface{} {
        return cache.NewManager(p.config)
    })
    return nil
}

func (p *CacheProvider) Boot(app foundation.Application) error {
    // Initialize cache connections
    return nil
}
```

### 3. Configuration

```go
import "github.com/donnigundala/dg-core/config"

// Load configuration
config.Load("config/app.yaml")

// Get values
dbHost := config.GetString("database.host", "localhost")
dbPort := config.GetInt("database.port", 5432)

// Inject into struct
var dbConfig DatabaseConfig
config.Inject("database", &dbConfig)
```

### 4. HTTP Routing

```go
import "github.com/donnigundala/dg-core/http"

router := http.NewRouter()

// Define routes
router.Get("/users", listUsers)
router.Post("/users", createUser)
router.Get("/users/:id", getUser)
router.Put("/users/:id", updateUser)
router.Delete("/users/:id", deleteUser)

// Use middleware
router.Use(middleware.Logger())
router.Use(middleware.Recovery())
router.Use(middleware.CORS())
```

### 5. Error Handling

```go
import "github.com/donnigundala/dg-core/errors"

// Create error
err := errors.New("user not found").
    WithCode("USER_NOT_FOUND").
    WithStatus(404).
    WithField("user_id", userID)

// Wrap error
err := errors.Wrap(dbErr, "failed to fetch user").
    WithCode("DATABASE_ERROR").
    WithStatus(500)

// Write HTTP error
errors.WriteHTTPError(w, err)
```

---

## Project Structure

```
myapp/
├── main.go                 # Application entry point
├── config/
│   ├── app.yaml           # Application configuration
│   └── database.yaml      # Database configuration
├── app/
│   ├── providers/         # Service providers
│   │   ├── app_provider.go
│   │   ├── database_provider.go
│   │   └── cache_provider.go
│   ├── http/
│   │   ├── controllers/   # HTTP controllers
│   │   ├── middleware/    # Custom middleware
│   │   └── routes.go      # Route definitions
│   └── models/            # Domain models
├── database/
│   └── migrations/        # Database migrations
└── storage/
    └── logs/              # Application logs
```

---

## Examples

### Complete Web Application

See the [skeleton project](https://github.com/donnigundala/dg-framework/tree/main/skeleton) for a complete example with:
- Service providers (cache, queue, redis)
- HTTP routes and controllers
- Configuration management
- Graceful shutdown
- Production deployment

### Custom Service Provider

```go
package providers

import (
    "github.com/donnigundala/dg-core/contracts/foundation"
    "github.com/donnigundala/dg-core/config"
)

type DatabaseProvider struct{}

func (p *DatabaseProvider) Register(app foundation.Application) error {
    app.Singleton("db", func() interface{} {
        var cfg DatabaseConfig
        config.Inject("database", &cfg)
        
        db, err := sql.Open(cfg.Driver, cfg.DSN)
        if err != nil {
            panic(err)
        }
        return db
    })
    return nil
}

func (p *DatabaseProvider) Boot(app foundation.Application) error {
    db, _ := app.Make("db").(*sql.DB)
    return db.Ping()
}
```

### Plugin with Metadata

```go
type MetricsPlugin struct {
    config MetricsConfig
}

func (p *MetricsPlugin) Name() string {
    return "metrics"
}

func (p *MetricsPlugin) Version() string {
    return "1.0.0"
}

func (p *MetricsPlugin) Dependencies() []string {
    return []string{"http"}
}

func (p *MetricsPlugin) Register(app foundation.Application) error {
    // Register metrics collector
    return nil
}

func (p *MetricsPlugin) Boot(app foundation.Application) error {
    // Start metrics server
    return nil
}
```

---

## Documentation

### Guides
- [Framework Philosophy](docs/PHILOSOPHY.md) - Design principles and rationale
- [Architecture Overview](docs/ARCHITECTURE.md) - System design and components
- [Service Providers](docs/PROVIDERS.md) - Creating and using providers
- [Plugin Development](docs/PLUGINS.md) - Building plugins

### API Reference
- [Container](https://pkg.go.dev/github.com/donnigundala/dg-core/container)
- [Configuration](https://pkg.go.dev/github.com/donnigundala/dg-core/config)
- [Foundation](https://pkg.go.dev/github.com/donnigundala/dg-core/foundation)
- [HTTP](https://pkg.go.dev/github.com/donnigundala/dg-core/http)
- [Errors](https://pkg.go.dev/github.com/donnigundala/dg-core/errors)

### Examples
- [Skeleton Project](https://github.com/donnigundala/dg-framework/tree/main/skeleton)
- [Provider Examples](examples/providers/)
- [HTTP Examples](examples/http/)

---

## Ecosystem

### Official Packages
- [dg-cache](https://github.com/donnigundala/dg-cache) - Multi-driver caching (Redis, Memory)
- [dg-queue](https://github.com/donnigundala/dg-queue) - Background job processing
- [dg-database](https://github.com/donnigundala/dg-database) - Database abstraction *(planned)*

### Community Packages
- Coming soon! Build your own and share with the community.

---

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./container -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

- **Container:** 89.7%
- **Config:** 73.5%
- **Middleware:** 66.7%
- **Errors:** 85.7%
- **Context Utils:** 95.9%
- **Overall:** ~80%

---

## Performance

dg-core is built on Gin, one of the fastest Go web frameworks:

```
BenchmarkRouter-8        5000000    250 ns/op    0 B/op    0 allocs/op
```

The DI container adds minimal overhead:
- Singleton resolution: ~50ns
- Transient resolution: ~100ns

---

## Requirements

- Go 1.24 or higher
- No OS-specific dependencies

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code of conduct
- Development setup
- Testing guidelines
- Pull request process

---

## Roadmap

### v1.x (Current)
- ✅ Core foundation (DI, config, providers)
- ✅ HTTP router and middleware
- ✅ Comprehensive testing (80%+ coverage)
- ✅ Documentation and guides
- 🔄 Ecosystem packages (cache, queue, database)

### v2.0 (Future)
- Plugin discovery and auto-loading
- Enhanced lifecycle hooks
- Middleware extraction (optional)
- Performance optimizations
- CLI tooling

---

## License

MIT License - see [LICENSE](LICENSE) for details

---

## Credits

Built with ❤️ by [Donni Gundala](https://github.com/donnigundala)

Inspired by:
- [Laravel](https://laravel.com/) - Service provider pattern
- [Gin](https://gin-gonic.com/) - HTTP performance
- [Spring Boot](https://spring.io/projects/spring-boot) - DI architecture

---

## Support

- 📖 [Documentation](docs/)
- 💬 [GitHub Discussions](https://github.com/donnigundala/dg-framework/discussions)
- 🐛 [Issue Tracker](https://github.com/donnigundala/dg-framework/issues)
- 📧 Email: donni@example.com

---

**Star ⭐ this repo if you find it helpful!**
