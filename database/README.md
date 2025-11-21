# `database` Package

## Overview

The `database` package provides a robust, production-grade framework for managing database connections. It is built on a modern, dependency-injection-friendly architecture that emphasizes context-awareness for superior performance and traceability.

The package offers a unified interface for multiple database backends, with built-in support for SQL (via GORM) and MongoDB.

## Core Concepts

### 1. The `Manager`

The `Manager` is the central component that handles the lifecycle of all named database connections. It is initialized from your application's configuration and provides a single point of access to all database providers.

### 2. The `Provider` Interface

All database drivers (SQL, MongoDB, etc.) implement a common `Provider` interface. This allows for consistent handling of different database systems. The package also provides extended interfaces like `SQLProvider` and `MongoProvider` for driver-specific functionality.

### 3. Context-Aware Operations

This is the most critical feature of the package. **All database operations should be performed with a `context.Context`**.

-   **Performance & Reliability**: Passing a context to database operations allows for graceful cancellation. If a user's request is canceled, any long-running database query associated with it will also be stopped, saving valuable database resources.
-   **End-to-End Tracing**: The database drivers are fully integrated with the `ctxutil` package. When you perform a database operation, the driver will use `ctxutil.LoggerFromContext(ctx)` to get a logger that includes the `request_id`. This enables you to trace a request from the HTTP server all the way down to the specific SQL query or MongoDB command that was executed.

#### For SQL (GORM):

The `SQLProvider` interface provides a special method:

-   **`GormWithContext(ctx context.Context) *gorm.DB`**: This is the **recommended** way to perform database operations. It returns a new GORM session that is bound to the request's context, ensuring all operations are traceable and cancellable.

## Full Usage Example

The following example demonstrates how to configure and use the database manager and its providers in a context-aware manner.

### 1. `config/app.yaml`

Define your database connections in your application's configuration file.

```yaml
databases:
  default_connection: "mysql_primary"
  connections:
    mysql_primary:
      driver: "sql"
      sql:
        driver_name: "mysql"
        primary:
          host: "127.0.0.1"
          port: "3306"
          user: "root"
          password: ${DB_PASSWORD} # Loaded from environment
          db_name: "my_app"
        log_level: "info"
        pool:
          max_open_conns: 25
          max_idle_conns: 5
          conn_max_lifetime: "1h"
    
    mongo_main:
      driver: "mongo"
      mongo:
        uri: ${MONGO_URI} # Loaded from environment
        database: "my_mongo_db"
```

### 2. `main.go` (or your application's entrypoint)

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/providers/database"
	"github.com/donnigundala/dgcore/ctxutil"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
}

func main() {
	// 1. Bootstrap logger and load configuration
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := config.Load("config/app.yaml"); err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Inject the database configurations into a struct
	var dbConfig database.Config
	if err := config.Inject("databases", &dbConfig); err != nil {
		logger.Error("failed to inject database configuration", "error", err)
		os.Exit(1)
	}

	// 3. Initialize the Database Manager
	dbManager, err := database.NewManager(dbConfig, database.WithLogger(logger))
	if err != nil {
		logger.Error("failed to initialize database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// 4. Get the default database connection
	dbProvider, err := dbManager.Connection() // Gets "mysql_primary"
	if err != nil {
		logger.Error("failed to get default database connection", "error", err)
		os.Exit(1)
	}

	// 5. Type-assert to the specific provider interface
	sqlProvider, ok := dbProvider.(database.SQLProvider)
	if !ok {
		logger.Error("default database is not a SQL provider")
		os.Exit(1)
	}

	// --- Performing Operations ---
	// Create a sample context, similar to what the server middleware would do.
	ctx := context.Background()
	ctx = ctxutil.WithLogger(ctx, logger.With("request_id", "xyz-789"))

	// 6. Get a context-aware GORM instance
	gormDB := sqlProvider.GormWithContext(ctx)

	// 7. Perform operations using the context-aware instance
	if err := gormDB.AutoMigrate(&User{}); err != nil {
		ctxutil.LoggerFromContext(ctx).Warn("Could not migrate User model", "error", err)
	}

	newUser := User{Name: "Context-Aware User"}
	if err := gormDB.Create(&newUser).Error; err != nil {
		ctxutil.LoggerFromContext(ctx).Error("Failed to create user", "error", err)
	} else {
		ctxutil.LoggerFromContext(ctx).Info("Created user", "name", newUser.Name, "id", newUser.ID)
	}
}
```
