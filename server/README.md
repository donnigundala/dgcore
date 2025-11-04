# `server/http` Package

## Overview

The `server/http` package provides a production-ready, configurable HTTP server for your application. It is built on Go's standard `net/http` package and enhanced with features for robustness, security, and ease of use.

## Features

- **Graceful Shutdown**: Handles OS signals (`SIGINT`, `SIGTERM`) to shut down cleanly, allowing in-flight requests to complete.
- **Functional Options Pattern**: Offers a clean, flexible API for programmatic configuration.
- **Framework-Integrated Configuration**: Seamlessly integrates with the `dgcore/config` package, allowing configuration via YAML files and environment variables.
- **Secure Defaults**: Enforces modern TLS versions and secure cipher suites when TLS is enabled.
- **Structured Logging**: Integrates with `slog` for consistent, structured logging.

## Configuration

The server is designed to be configured through your application's central configuration system. The framework automatically registers default values, which you can easily override.

The configuration loading follows this order of precedence (highest to lowest):
1.  **Environment Variables**
2.  **`config.yaml` File**
3.  **Framework Defaults**

### YAML Configuration

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
 
 | Key                 | Default     |
 | ------------------- | ----------- |
 | `host`              | `127.0.0.1` |
 | `port`              | `8080`      |
| `read_timeout`      | `5s`        |
 | `write_timeout`     | `10s`       |
 | `idle_timeout`      | `120s`      |
 | `tls.enabled`       | `false`     |
 | `tls.cert_file`     | `""`        |
 | `tls.key_file`      | `""`        |
 | `tls.tls_version`   | `TLS1.2`    |


## Usage
 
 The `example/main.go` file provides a complete, runnable example of how to bootstrap and run the server. The typical workflow is:
 
 1.  **Initialize a logger.**
 2.  **Load configuration** using `config.Load()`.
 3.  **Inject the server config** into the `dghttp.Config` struct using `config.Inject("server.http", &cfg)`.
 4.  **Create a new server** instance using `dghttp.NewServerFromConfig()`.
 5.  **Start the server** with `srv.Start()`.
 6.  **Block and wait for shutdown** with `srv.WaitForShutdown()`.

```go
 // See core/server/http/example/main.go for the full implementation.
 
 func main() {
     // ... logger and config setup ...
 
     var serverCfg dghttp.Config
     if err := config.Inject("server.http", &serverCfg); err != nil {
         // handle error
     }
 
     mux := http.NewServeMux()
     // ... define handlers ...
 
     srv := dghttp.NewServerFromConfig(&serverCfg, mux)
     srv.Start()
     srv.WaitForShutdown(10 * time.Second)
 }
 ```