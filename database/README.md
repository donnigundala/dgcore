# `core/database` Package

## Overview

The `database` package provides a robust, production-grade framework for managing database connections in Go applications. It is built on a modern, Dependency-Injection-friendly architecture that is consistent across the `dg-framework`.

The package offers a unified interface for multiple database backends, with built-in support for high availability, read/write splitting (for SQL), and advanced configuration through simple YAML files.

## Features

- **Dependency Injection (DI) Architecture**: Decoupled from global state. The `DatabaseManager` is created by injecting a configuration struct, making it highly testable and easy to integrate.
- **Unified Configuration**: Configure all database connections (Postgres, MySQL, SQLite, MongoDB) in a clean, structured YAML file. The library automatically handles overrides from environment variables.
- **High Availability & Read/Write Splitting (SQL)**: Automatically routes write operations to the primary node and read operations across replicas. (Note: Failover logic may need to be managed by the application or a higher-level component).
- **Advanced Connection Pooling**: Provides granular control over connection pool settings for each database.
- **Secure by Default**: Encourages loading sensitive credentials (passwords, URIs) from environment variables, which automatically override values in YAML files.
- **Structured Logging**: Integrates with Go's `log/slog` for structured, machine-readable logs.
- **Modular Providers**: Easily extensible to support new database types. Comes with built-in providers for SQL (via GORM) and MongoDB.

## Configuration

Configuration is handled via YAML files (e.g., `config/database.yaml`) loaded by your application's central configuration engine. The library defines clear Go structs (`database.ManagerConfig`) that you unmarshal your configuration into.

**The framework's configuration loader (Viper) automatically overrides YAML values with environment variables.** The environment variable name is constructed from the YAML path, in uppercase, with `.` replaced by `_`.

**Example:** The YAML key `connections.my_postgres.sql.primary.password` is overridden by the environment variable `DATABASES_CONNECTIONS_MY_POSTGRES_SQL_PRIMARY_PASSWORD`.

### Example `database.yaml`

```yaml
# your-app/config/database.yaml

# The top-level key 'databases' is used to inject the config.
databases:
  # Set the default connection to be used when no name is specified.
  default_connection: "my_sqlite"

  # Define all your database connections here.
  connections:
    # --- SQLite Example ---
    my_sqlite:
      driver: "sql"
      sql:
        driver_name: "sqlite"
        primary:
          db_name: "example.db"
        log_level: "info"

    # --- PostgreSQL Example ---
    my_postgres:
      driver: "sql"
      policy:
        ping_interval: "15s"
      sql:
        driver_name: "postgres"
        primary:
          host: "localhost"
          port: "5432"
          user: "postgres"
          password: "password" # Default, overridden by env var in production
          db_name: "my_app"
        pool:
          max_open_conns: 20
        log_level: "warn"

    # --- MongoDB Example ---
    my_mongo:
      driver: "mongo"
      policy:
        ping_interval: "20s"
      mongo:
        uri: "mongodb://localhost:27017" # Default, overridden by env var in production
        database: "my_mongo_db"
        pool:
          max_pool_size: 50
        log_level: "info"
```

## Usage

The idiomatic workflow involves loading configuration into a struct and injecting it into the `DatabaseManager` constructor.

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/database"
	"gorm.io/gorm"
)

// Example User model
type User struct {
	gorm.Model
	Name string
}

func main() {
	// 1. Initialize Application Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 2. Load Configuration
	// Assumes a loader that reads all .yaml files from ./config and handles env vars.
	config.Load()

	// 3. Inject Configuration into Struct
	// Unmarshal the 'databases' section from the loaded config into our struct.
	var dbManagerConfig database.ManagerConfig
	if err := config.Inject("databases", &dbManagerConfig); err != nil {
		slog.Error("Failed to inject database configuration", "error", err)
		os.Exit(1)
	}

	// 4. Create the DatabaseManager with Dependency Injection
	dbManager, err := database.NewManager(dbManagerConfig, database.WithLogger(logger))
	if err != nil {
		slog.Error("Failed to create database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	slog.Info("Database manager initialized successfully.")

	// 5. Get and Use a Database Connection
	// Get the default connection as specified in the YAML.
	db, err := dbManager.Connection()
	if err != nil {
		slog.Error("Failed to get default database connection", "error", err)
		os.Exit(1)
	}

	// We need to type-assert the provider to get the specific GORM instance.
	sqlDB, ok := db.(database.SQLProvider)
	if !ok {
		slog.Error("Default database is not a SQL provider")
		os.Exit(1)
	}
	gormDB := sqlDB.Gorm().(*gorm.DB)

	// 6. Perform Operations
	slog.Info("--- Performing DB Operations ---")
	if err := gormDB.AutoMigrate(&User{}); err != nil {
		slog.Warn("Could not migrate User model", "error", err)
	}

	// Write
	newUser := User{Name: "Framework User"}
	if err := gormDB.Create(&newUser).Error; err != nil {
		slog.Error("Failed to create user", "error", err)
	} else {
		slog.Info("Created user", "name", newUser.Name, "id", newUser.ID)
	}

	// Read
	var fetchedUser User
	if err := gormDB.First(&fetchedUser, newUser.ID).Error; err != nil {
		slog.Error("Failed to read user", "error", err)
	} else {
		slog.Info("Fetched user", "name", fetchedUser.Name, "id", fetchedUser.ID)
	}
}
```
