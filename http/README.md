# `server` Package

## Overview

The `server` package provides a production-grade, robust, and flexible framework for managing the lifecycle of multiple servers (e.g., HTTP, gRPC) within a single application. It is built with modern Go practices, focusing on graceful shutdown, structured logging, and context-aware request handling.

## Core Concepts

### 1. The `Manager`

The `Manager` is the heart of the package. It orchestrates the startup and graceful shutdown of all registered servers.

-   **Centralized Control**: Start and stop multiple servers (e.g., a public HTTP API and a private gRPC service) from a single point.
-   **Graceful Shutdown**: Listens for OS signals (`SIGINT`, `SIGTERM`) and ensures all servers shut down cleanly, allowing in-flight requests to complete.
-   **Robust Error Handling**: Captures and returns errors from both server startup and shutdown processes.

### 2. The `Runnable` Interface

Any component that can be started and stopped can be managed by the `Manager`, as long as it implements the simple `Runnable` interface:

```go
type Runnable interface {
    Start() error
    Shutdown(ctx context.Context) error
}
```

This design makes the manager highly extensible. The package provides a built-in `HTTPServer` that implements this interface.

### 3. Context-Aware Logging & Tracing

For effective debugging and tracing in a production environment, it's crucial to track individual requests. The `server` package provides a middleware for this purpose.

-   **`RequestIDMiddleware`**: An `http.Handler` middleware that automatically injects a unique `request_id` into every incoming request's context.
-   **`LoggerFromContext(ctx)`**: A helper function that retrieves a context-aware `slog.Logger` from the request context. This logger will automatically include the `request_id` in all its log entries.

This allows you to trace the entire lifecycle of a single request across different functions and packages.

### 4. Centralized Error Handling

To avoid repetitive error-handling boilerplate in your HTTP handlers, the package provides a custom `HandlerFunc` type.

-   **`server.HandlerFunc`**: A function signature of `func(w http.ResponseWriter, r *http.Request) error`.
-   **Automatic Error Handling**: When you wrap your handler with `server.HandlerFunc`, any error returned will be automatically caught, logged (with its `request_id`), and a generic `500 Internal Server Error` will be sent to the client.

## Full Usage Example

The following example demonstrates how all these components work together. It is based on the code in `server/example/main.go`.

### 1. `config/server.yaml`

First, define your server configuration:

```yaml
server:
  http:
    addr: ":8080"
    read_timeout: 5s
    write_timeout: 10s
    idle_timeout: 120s
```

### 2. `main.go`

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/donnigundala/dg-core/config"
	"github.com/donnigundala/dg-core/server"
)

func main() {
	// 1. Initialize a global logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 2. Load configuration from file
	if err := config.Load("config/server.yaml"); err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 3. Inject the 'server.http' section into a struct
	var serverCfg server.Config
	if err := config.Inject("server.http", &serverCfg); err != nil {
		logger.Error("failed to inject server configuration", "error", err)
		os.Exit(1)
	}

	// 4. Create a server Manager
	mgr := server.NewManager(
		server.WithLogger(logger),
		server.WithShutdownTimeout(20*time.Second),
	)

	// 5. Define handlers and wrap them with middleware
	mux := http.NewServeMux()
	mux.Handle("/hello", server.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Get the logger from the context to include the request_id
		log := server.LoggerFromContext(r.Context())
		log.Info("Handling /hello request")
		
		// You can return errors directly
		if r.URL.Query().Get("fail") == "true" {
			return errors.New("a simulated error occurred")
		}
		
		fmt.Fprintln(w, "Hello, World!")
		return nil
	}))

	// Wrap the main router with the RequestIDMiddleware
	var httpHandler http.Handler = mux
	httpHandler = server.RequestIDMiddleware(httpHandler)

	// 6. Create and register the HTTP server
	httpServer := server.NewHTTPServer(serverCfg, httpHandler)
	mgr.Register("http-public", httpServer)

	// 7. Start the application and wait for a shutdown signal
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("starting server manager")
	if err := mgr.RunAll(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("server manager failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server manager stopped gracefully")
}
```
