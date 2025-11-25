# dg-core Framework Philosophy

## Core Principle: Pragmatic Minimalism

dg-core follows a **"pragmatic minimalism"** approach — we keep the core small and focused, but include commonly-needed features to enable rapid development without sacrificing modularity.

---

## What Belongs in the Core?

### ✅ Essential Foundation (Always Included)

These components are fundamental to any application built with dg-core:

1. **Dependency Injection Container** (`container/`)
   - Enables loose coupling and testability
   - Foundation for the plugin system
   - ~300 lines of code

2. **Configuration Management** (`config/`)
   - Environment-based configuration
   - YAML/ENV file support
   - Essential for all applications
   - ~500 lines of code

3. **Application Foundation** (`foundation/`)
   - Application lifecycle management
   - Service provider system
   - Graceful shutdown handling
   - ~400 lines of code

4. **Error Handling** (`errors/`)
   - Standardized error patterns
   - HTTP error conversion
   - Stack trace support
   - ~200 lines of code

5. **Context Utilities** (`ctxutil/`)
   - Request ID generation
   - Logger storage
   - Request metadata extraction
   - ~300 lines of code

6. **Contracts/Interfaces** (`contracts/`)
   - Framework contracts
   - Enable plugin development
   - ~200 lines of code

7. **Logging Abstraction** (`logging/`)
   - Minimal wrapper around Go's slog
   - ~100 lines of code

8. **Testing Utilities** (`testing/`)
   - Framework testing helpers
   - ~200 lines of code

**Total Core Essentials:** ~2,200 lines (34% of codebase)

---

### ⚡ Common Features (Batteries Included)

These features are included because they're needed by the vast majority of applications:

1. **HTTP Router** (`http/router.go`)
   - 90% of Go applications are web services
   - Provides consistent, framework-integrated API
   - Wraps Gin for performance and ecosystem
   - Can be ignored for CLI/worker applications

2. **HTTP Middleware** (`http/middleware/`)
   - Common patterns every web app needs
   - CORS, logging, recovery, security, rate limiting
   - Integrated with framework's error handling and logging
   - ~2,000 lines of code

3. **Validation** (`validation/`)
   - Request validation for HTTP endpoints
   - Custom validator support
   - Tight integration with HTTP layer
   - ~800 lines of code

**Total Common Features:** ~3,000 lines (46% of codebase)

---

### ❌ Not in Core (Separate Packages)

These features are application-specific and belong in separate packages:

- **Database** → `dg-database` (planned)
- **Cache** → `dg-cache` ✅ (released)
- **Queue** → `dg-queue` ✅ (released)
- **Email** → `dg-mail` (planned)
- **Storage** → `dg-storage` (planned)
- **Authentication** → `dg-auth` (planned)

---

## Why Include HTTP in Core?

### The Pragmatic Decision

We include HTTP router and middleware in dg-core for several reasons:

1. **90% Use Case**
   - The vast majority of Go applications are web services
   - HTTP is the primary interface for modern applications
   - Including it reduces friction for the common case

2. **Framework Integration**
   - HTTP middleware integrates with framework's error handling
   - Uses framework's logging and context utilities
   - Provides consistent patterns across the framework

3. **Rapid Development**
   - Developers can start building immediately
   - No need to install and configure separate HTTP package
   - Reduces decision fatigue for beginners

4. **Industry Precedent**
   - Most successful frameworks include HTTP (Rails, Django, NestJS)
   - Go-specific frameworks (Gin, Echo, Fiber) are HTTP-first
   - Modular frameworks often struggle with adoption

### For Non-HTTP Applications

If you're building a CLI tool, worker, or other non-HTTP application:

- Simply don't use the `http` package
- The HTTP dependency (Gin) won't affect your binary if unused
- All core features (container, config, providers) work independently

---

## Comparison with Other Frameworks

### Laravel (PHP)
- **Core:** Container, service providers, facades
- **Separate:** HTTP, routing, validation, database
- **Our approach:** Similar, but include HTTP for pragmatism

### Spring Boot (Java)
- **Core:** Dependency injection, application context
- **Separate:** Web, validation, data access
- **Our approach:** Similar philosophy, different execution

### NestJS (TypeScript)
- **Core:** DI, modules, lifecycle
- **Separate:** Platform adapters (Express/Fastify)
- **Our approach:** More batteries-included

### Go Frameworks
- **Gin/Echo/Fiber:** Monolithic (all-in-one)
- **Buffalo:** Modular but complex
- **Our approach:** Balanced middle ground

---

## Design Principles

### 1. Convention over Configuration
- Sensible defaults for common use cases
- Override when needed, but works out of the box

### 2. Explicit over Implicit
- Clear service provider registration
- Visible dependency injection
- No magic, no reflection where avoidable

### 3. Modular Architecture
- Core features are loosely coupled
- Plugin system enables extensibility
- Use only what you need

### 4. Developer Experience
- Familiar patterns (Laravel-inspired)
- Comprehensive documentation
- Clear error messages

### 5. Production Ready
- Performance (via Gin)
- Graceful shutdown
- Structured logging
- Error handling

---

## Future Direction

### v1.x (Current)
- ✅ Maintain current structure
- ✅ Focus on stability and documentation
- ✅ Build ecosystem (dg-database, dg-auth, etc.)
- ✅ Gather community feedback

### v2.0 (Future Consideration)
- ⚠️ Consider extracting middleware to `dg-http-middleware`
- ⚠️ Keep minimal router in core
- ⚠️ Reduce core to ~3,500 lines (46% smaller)
- ⚠️ Maintain backward compatibility where possible

**Decision Point:** We'll evaluate based on:
- Community feedback
- Real-world usage patterns
- Ecosystem maturity
- Breaking change impact

---

## Package Size Guidelines

We aim to keep packages focused and maintainable:

- **Core packages:** < 500 lines each
- **Feature packages:** < 1,000 lines each
- **Total core:** < 4,000 lines (currently ~6,500)
- **Well-documented:** Every public API has godoc

---

## Plugin System Philosophy

### Extensibility over Bloat

Instead of adding every feature to core, we provide:

1. **Service Provider Pattern**
   - Register and boot lifecycle
   - Dependency injection integration
   - Clean separation of concerns

2. **Plugin Metadata**
   - Name, version, dependencies
   - Discovery and validation
   - Optional for simple providers

3. **Ecosystem Packages**
   - dg-cache, dg-queue (released)
   - dg-database, dg-auth (planned)
   - Community plugins welcome

### Example: Why Cache is Separate

Cache is **not** in core because:
- Not every application needs caching
- Multiple driver options (Redis, Memcached, Memory)
- Adds significant dependencies
- Can be added via service provider

But it's **easy to add**:
```go
import "github.com/donnigundala/dg-cache"

app.Register(providers.NewCacheServiceProvider(cacheConfig))
```

---

## Measuring Success

We consider dg-core successful when:

1. **Adoption**
   - Developers choose it for new projects
   - Existing projects migrate to it
   - Community contributions grow

2. **Productivity**
   - Faster time-to-production
   - Less boilerplate code
   - Clear upgrade paths

3. **Maintainability**
   - Core stays focused and small
   - Ecosystem grows organically
   - Breaking changes are rare

4. **Performance**
   - Competitive with Gin/Echo
   - Minimal overhead from abstractions
   - Efficient resource usage

---

## Questions & Answers

### "Why not just use Gin directly?"

dg-core adds:
- Dependency injection container
- Service provider pattern
- Standardized error handling
- Configuration management
- Plugin ecosystem
- Consistent patterns across features

### "Why not make everything a plugin?"

Balance:
- Too minimal = high friction for common cases
- Too bloated = defeats purpose of modularity
- Our approach = pragmatic middle ground

### "Will you extract HTTP in v2.0?"

Maybe:
- We'll evaluate based on real-world usage
- Community feedback is crucial
- Backward compatibility is important
- No decision made yet

### "How do I contribute?"

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code style guidelines
- Testing requirements
- Pull request process
- Community standards

---

## Summary

**dg-core is pragmatically minimal:**
- Small core (~2,200 lines) with essential features
- Batteries included (~3,000 lines) for common use cases
- Extensible via service providers and plugins
- Focused on developer productivity and production readiness

**We believe this balance:**
- Reduces friction for the 90% use case
- Maintains modularity for the 10% edge cases
- Enables rapid development without sacrificing flexibility
- Positions dg-core as a practical choice for Go web development

---

**Philosophy Version:** 1.0  
**Last Updated:** 2025-11-25  
**Feedback:** [GitHub Issues](https://github.com/donnigundala/dg-framework/issues)
