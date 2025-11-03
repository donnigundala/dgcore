# `core/database` Package

## Overview

The `database` package provides a robust, production-grade framework for managing database connections in Go applications. It offers a unified interface for multiple database backends, with built-in support for high availability, failover, read/write splitting, and advanced configuration.

## Features

- **High Availability & Failover**: Automatically detects primary node failures and seamlessly fails over to healthy replicas. Includes a background worker that attempts to reconnect to the primary with exponential backoff.
- **Read/Write Splitting**: Intelligently routes queries to distribute load. Write operations are sent to the primary, while read operations are distributed across replicas in a round-robin fashion. This is enabled automatically when replicas are configured.
- **Advanced Connection Pooling**: Provides granular control over connection pool settings for each database, including max/min connections and connection lifetimes, allowing for fine-tuning of performance.
- **Security Hardening**:
    - **Secrets Management**: Supports loading sensitive credentials (DSNs, URIs, passwords) from environment variables to avoid storing them in plaintext configuration.
    - **TLS/SSL Enforcement**: Enforces encrypted connections and supports custom Certificate Authorities (CAs) and client certificates.
- **Metrics & Observability**: Features a flexible `MetricsProvider` interface to integrate with monitoring systems like Prometheus. It exports key metrics on connection health, failover events, and pool statistics.
- **Structured Logging**: Integrates with Go's `log/slog` for structured, machine-readable logs, enhancing debuggability and observability. GORM logs are also routed through `slog`, with context-aware `trace_id` injection.
- **Framework-Idiomatic Configuration**: Integrates seamlessly with a framework-style configuration system, loading settings from a central configuration block.
- **Multi-Engine SQL Support**: Easily switch between PostgreSQL, MySQL, SQLite, and other SQL databases via configuration, using parameter-based connection details instead of raw DSNs.
- **Modular Providers**: Easily extensible to support new database types. Comes with built-in providers for SQL (PostgreSQL, MySQL, SQLite) and MongoDB.

## Configuration

The package is designed to be configured via a central configuration file (e.g., `config.yaml`) that is loaded by the application. The library provides a template file, `database_config.go`, which should be copied into your application's `config` directory to register the default settings.

### Example YAML Configuration

```yaml
# your-app/config/config.yaml

databases:
  default_trace_id_key: X-Trace-ID # Optional: Key to extract trace ID from context for logging

  my_postgres:
    driver: sql
    policy:
      ping_interval: 10s
      max_failures: 3
    sql:
      driver_name: postgres # Specifies the SQL driver to use
      primary:
        host: {from_env: POSTGRES_PRIMARY_HOST}
        port: {from_env: POSTGRES_PRIMARY_PORT}
        user: {from_env: POSTGRES_PRIMARY_USER}
        password: {from_env: POSTGRES_PRIMARY_PASSWORD}
        db_name: {from_env: POSTGRES_PRIMARY_DBNAME}
        params: # Optional: additional driver-specific parameters
          sslmode: disable
      replicas:
        - host: {from_env: POSTGRES_REPLICA_1_HOST}
          port: {from_env: POSTGRES_REPLICA_1_PORT}
          user: {from_env: POSTGRES_REPLICA_1_USER}
          password: {from_env: POSTGRES_REPLICA_1_PASSWORD}
          db_name: {from_env: POSTGRES_REPLICA_1_DBNAME}
          params:
            sslmode: disable
      pool:
        max_open_conns: 25
        max_idle_conns: 10
      log_level: warn # GORM log level (silent, error, warn, info)

  my_mysql:
    driver: sql
    sql:
      driver_name: mysql
      primary:
        host: {from_env: MYSQL_PRIMARY_HOST}
        port: {from_env: MYSQL_PRIMARY_PORT}
        user: {from_env: MYSQL_PRIMARY_USER}
        password: {from_env: MYSQL_PRIMARY_PASSWORD}
        db_name: {from_env: MYSQL_PRIMARY_DBNAME}
      pool:
        max_open_conns: 25
        max_idle_conns: 10
      log_level: warn

  my_sqlite:
    driver: sql
    sql:
      driver_name: sqlite
      primary:
        db_name: {from_env: SQLITE_DBNAME} # SQLite uses file path as db_name
      log_level: warn

  my_mongo:
    driver: mongo
    mongo:
      primary_uri:
        from_env: MONGO_PRIMARY_URI
      database: "my_app_db"
      pool:
        max_pool_size: 50
      log_level: info # Mongo driver log level (debug, info, warn, error)
```

## Usage

Below is an example of the idiomatic workflow for using the `database` package within a larger application.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/donnigundala/dgcore/database"
	"github.com/donnigundala/dgcore/database/contracts"
	"github.com/donnigundala/dgframework/config" // Assuming your framework's config package
	"gorm.io/gorm"
)

// Define a dummy model for demonstration
type User struct {
	gorm.Model
	Name string
}

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	TraceIDContextKey ContextKey = "X-Trace-ID"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- 1. Initialize Application Logger ---
	// In a real app, this would be configured based on environment (e.g., JSON for production)
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(appSlog) // Set as default for convenience

	// --- 2. Inject Configuration ---
	// This step is typically handled by your application's startup logic.
	var dbConfigs map[string]*contracts.Config
	if err := config.Inject("databases", &dbConfigs); err != nil {
		log.Fatalf("Failed to inject database configurations: %v", err)
	}

	// --- 3. Register Configurations with the Manager ---
	log.Println("Registering database configurations...")
	mgr := database.Manager()
	// Set the application's slog logger on the manager. This logger will be passed to providers.
	mgr.SetLogger(appSlog)

	for name, cfg := range dbConfigs {
		// Assign the trace ID key from the config (e.g., from YAML) to each config
		// If not set in config, it defaults to empty, meaning no trace ID will be extracted.
		// For this example, we explicitly set it to our local ContextKey.
		cfg.TraceIDKey = string(TraceIDContextKey)
		mgr.Register(name, cfg)
	}

	// --- 4. Connect to All Databases ---
	log.Println("Connecting to all registered databases...")
	// Create a context with a trace ID for the connection phase
	connCtx := context.WithValue(ctx, TraceIDContextKey, "conn-trace-123")
	if err := mgr.ConnectAll(connCtx); err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer mgr.Close()

	// --- 5. Health Check ---
	log.Println("Performing initial health check...")
	// Use a context with a different trace ID for the health check
	healthCtx := context.WithValue(ctx, TraceIDContextKey, "health-trace-456")
	mgr.PrintHealth(healthCtx)

	// --- 6. Usage ---
	log.Println("\n--- Using the Database Provider ---")
	pgProvider := mgr.Get("my_postgres") // Using the name from the config
	if pgProvider == nil {
		log.Fatal("Postgres provider 'my_postgres' not found")
	}

	// Create a context with a trace ID for database operations
	opCtx := context.WithValue(ctx, TraceIDContextKey, "op-trace-789")

	// Get a connection for write operations
	log.Println("Getting a writer connection...")
	writerDB, ok := pgProvider.GetWriter().(*gorm.DB)
	if !ok || writerDB == nil {
		log.Fatal("Failed to get a valid writer connection")
	}
	fmt.Println("Successfully got a writer connection.")

	// Perform a simple write operation using the context
	log.Println("\n--- Performing DB Operations ---")
	if err := writerDB.WithContext(opCtx).AutoMigrate(&User{}); err != nil {
		log.Printf("Warning: Could not migrate User model: %v", err)
	}

	// Write
	newUser := User{Name: "Framework User"}
	if err := writerDB.WithContext(opCtx).Create(&newUser).Error; err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	log.Printf("Created user: %s (ID: %d)", newUser.Name, newUser.ID)

	// Get a reader for SELECT operations (will use replicas if available)
	log.Println("Getting a reader connection...")
	readerDB, ok := pgProvider.GetReader().(*gorm.DB)
	if !ok || readerDB == nil {
		log.Fatal("Failed to get a valid reader connection")
	}
	fmt.Println("Successfully got a reader connection.")

	// Read
	var fetchedUser User
	if err := readerDB.WithContext(opCtx).First(&fetchedUser, newUser.ID).Error; err != nil {
		log.Fatalf("Failed to read user: %v", err)
	}
	log.Printf("Fetched user: %s (ID: %d)", fetchedUser.Name, fetchedUser.ID)

	// --- Keep application running ---
	log.Println("\n--- Application is running. --- ")
	select {
	case <-time.After(30 * time.Second):
		log.Println("Example finished.")
	case <-ctx.Done():
		log.Println("Context cancelled.")
	}
}

// Helper to check for environment variables
func init() {
	// Check for environment variables needed by the example's default config
	if os.Getenv("POSTGRES_PRIMARY_HOST") == "" {
		log.Println("Warning: POSTGRES_PRIMARY_HOST environment variable is not set.")
	}
	// Add checks for other necessary environment variables if needed
}
