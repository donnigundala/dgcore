# `server` Package

## Overview

The `server` package provides a robust and flexible framework for managing the lifecycle of multiple servers (e.g., HTTP, gRPC) within an application. It introduces a `Manager` that orchestrates the starting, graceful shutdown, and health monitoring of various `Runnable` server instances.

## Features

-   **Centralized Server Management**: Start, stop, and manage multiple server types from a single point of control.
-   **Graceful Shutdown**: Implements graceful shutdown for all registered servers, allowing in-flight requests to complete before termination.
-   **Configurable Shutdown Timeout**: Allows setting a custom timeout for the graceful shutdown process.
-   **Structured Logging**: Integrates with `slog` for consistent, structured logging across all server operations.
-   **Extensible**: Designed with the `Runnable` interface, making it easy to integrate any server type (HTTP, gRPC, custom) that implements the interface.

## `Manager` Usage

The `Manager` is the core component for orchestrating your servers.

### Creating a Manager

You can create a new `Manager` instance and configure it using functional options:

```go
import (
	"log/slog"
	"os"
	"time"

	"github.com/donni/dg-framework/core/server"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

    mgr := server.NewManager(
        server.WithLogger(logger),
        server.WithShutdownTimeout(10*time.Second),
    )
    // ... register and run servers
}
```

### Registering Servers

Any server that implements the `server.Runnable` interface can be registered with the `Manager`.

```go
// Assuming httpServer is an instance of a server.Runnable
mgr.Register("http-public", httpServer)
```

### Running Servers

You can run all enabled servers or specific servers by name. The `RunAll` and `Run` methods are blocking and will manage the server lifecycle, including graceful shutdown upon context cancellation (e.g., from OS signals).

```go
import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
    // ... manager setup ...

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Listen for OS shutdown signals
    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        logger.Info("Shutdown signal received")
        cancel() // Trigger context cancellation
    }()

    logger.Info("Starting server manager...")
    if err := mgr.RunAll(ctx); err != nil {
        logger.Error("Server manager failed", "error", err)
        os.Exit(1)
    }

    logger.Info("Server manager stopped gracefully.")
}
```

## `server/http` Package

### Overview

The `server/http` package provides a production-ready, configurable HTTP server for your application. It is built on Go's standard `net/http` package and enhanced with features for robustness, security, and ease of use.

### Features

-   **Graceful Shutdown**: Handles OS signals (`SIGINT`, `SIGTERM`) to shut down cleanly, allowing in-flight requests to complete.
-   **Functional Options Pattern**: Offers a clean, flexible API for programmatic configuration.
-   **Framework-Integrated Configuration**: Seamlessly integrates with the `dgcore/config` package, allowing configuration via YAML files and environment variables.
-   **Secure Defaults**: Enforces modern TLS versions and secure cipher suites when TLS is enabled.
-   **Structured Logging**: Integrates with `slog` for consistent, structured logging.

## Configuration

The server is designed to be configured through your application's central configuration system. The framework automatically registers default values, which you can easily override.

The configuration loading follows this order of precedence (highest to lowest):
1.  **Environment Variables**
2.  **`config.yaml` File**
3.  **Framework Defaults**

### YAML Configuration Example

```yaml
server:
  http:
    addr: ":8080"
    read_timeout: 5s
    write_timeout: 10s
    idle_timeout: 120s
    # tls:
    #   enabled: true
    #   cert_file: "/etc/ssl/certs/server.crt"
    #   key_file: "/etc/ssl/private/server.key"
    #   tls_version: "TLS1.3"
```

### Environment Variables Example

```bash
# Timeouts
export SERVER_HTTP_READ_TIMEOUT="10s"
export SERVER_HTTP_WRITE_TIMEOUT="15s"
export SERVER_HTTP_IDLE_TIMEOUT="3m"

# TLS settings
export SERVER_HTTP_TLS_ENABLED=true
export SERVER_HTTP_TLS_CERT_FILE="/etc/ssl/certs/server.crt"
export SERVER_HTTP_TLS_KEY_FILE="/etc/ssl/private/server.key"
export SERVER_HTTP_TLS_TLS_VERSION="TLS1.3"
```

### Default Values

If no overrides are provided, the server will use these sensible defaults:

| Key               | Default   |
| :---------------- | :-------- |
| `addr`            | `:8080`   |
| `read_timeout`    | `5s`      |
| `write_timeout`   | `10s`     |
| `idle_timeout`    | `120s`    |
| `tls.enabled`     | `false`   |
| `tls.cert_file`   | `""`      |
| `tls.key_file`    | `""`      |
| `tls.tls_version` | `TLS1.2`  |

## Usage Example

The `core/server/example/example.go` file provides a complete, runnable example of how to bootstrap and run an HTTP server using the `Manager`.

```go
// See core/server/example/example.go for the full implementation.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/donni/dg-framework/core/config"
	"github.com/donni/dg-framework/core/server"
	"github.com/donni/dg-framework/core/server/http/handler"
)

type AppConfig struct {
	Server struct {
		HTTP server.Config `yaml:"http"`
	} `yaml:"server"`
}

func main() {
	// =========================================================================
	// Configuration
	// =========================================================================
	var appCfg AppConfig
	if err := config.Load("config/server.yaml", &appCfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// =========================================================================
	// Logger
	// =========================================================================
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// =========================================================================
	// Server Manager
	// =========================================================================
	mgr := server.NewManager(
		server.WithLogger(logger),
		server.WithShutdownTimeout(10*time.Second),
	)

	// =========================================================================
	// HTTP Server
	// =========================================================================
	// Create a simple router
	mux := http.NewServeMux()
	mux.Handle("/hello", handler.Wrapper(func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintln(w, "Hello, World!")
		return nil
	}))

	// Create the HTTP server using the config
	httpServer := server.NewHTTPServer(
		appCfg.Server.HTTP,
		server.WithHTTPHandler(mux),
	)

	// Register the server with the manager
	mgr.Register("http-public", httpServer)

	// =========================================================================
	// Start Servers
	// =========================================================================
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for shutdown signals
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("Shutdown signal received")
		cancel()
	}()

	logger.Info("Starting server manager...")
	if err := mgr.RunAll(ctx); err != nil {
		logger.Error("Server manager failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server manager stopped gracefully.")
}
```
