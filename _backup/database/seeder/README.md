# `core/seeder` Package

## Overview

The `seeder` package provides a clean, robust, and testable mechanism for populating a database with initial data. It is designed with a "dependency-injection first" approach, avoiding global state and `init()`-based magic in favor of explicit, clear, and controllable execution.

## Features

- **Dependency-Injected**: The seeder is instantiated with its dependencies (`*gorm.DB`, `*slog.Logger`), making it easy to use and test in isolation.
- **Explicit Registration**: Seeders are registered programmatically, making the execution flow clear and preventing "spooky action at a distance."
- **Ordered Execution**: You can define a precise execution order for seeders, which is critical for data with relational dependencies. If no order is set, seeders run alphabetically for deterministic behavior.
- **Transactional Execution**: Run all seeders within a single database transaction, ensuring that the entire seeding process succeeds or fails atomically.
- **Structured Logging**: Fully integrated with `slog` for consistent, structured logging that includes the component name (`seeder`).

## Why Explicit Registration?

This package intentionally avoids `init()`-based auto-registration. While auto-registration can seem convenient, it introduces several problems:

1.  **Hidden Behavior**: It's impossible to know which seeders will run just by looking at the main application logic.
2.  **Poor Testability**: Global registries create state that persists between tests, leading to flaky and non-isolated tests.
3.  **Lack of Control**: It's difficult to conditionally run different sets of seeders for different environments (e.g., testing vs. staging).

By requiring explicit registration in your application's entry point (the "composition root"), the `seeder` package ensures your code is **clear, testable, and flexible.**

## Usage

Using the seeder typically involves creating a dedicated command within your application.

### Step 1: Define Your Seeder Functions

A seeder is any function that matches the `dgseeder.SeedFunc` signature.

```go
// in my-app/database/seeders/user_seeder.go

package seeders

import (
    "gorm.io/gorm"
    "my-app/internal/models" // Your app's models
)

func UserSeeder(db *gorm.DB) error {
    users := []models.User{
        {Name: "Admin User", Email: "admin@example.com"},
        {Name: "Test User", Email: "test@example.com"},
    }
    // Using FirstOrCreate to prevent duplicates on re-runs
    return db.FirstOrCreate(&users).Error
}
```

### Step 2: Create a Seeder Command

In your application's `cmd` directory, create an entry point for seeding. This file will bootstrap the application, instantiate the seeder, register your functions, and run them.

```go
// in my-app/cmd/seed/main.go

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/providers/database"
	dgseeder "github.com/donnigundala/dgcore/providers/database/seeder"
	"gorm.io/gorm"

	// Import your seeder functions
	"my-app/database/seeders"
)

func main() {
	// 1. Bootstrap your application (logger, config, database)
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(appSlog)

	config.Load()

	// Inject the 'databases' section from config into the ManagerConfig struct.
	var dbManagerConfig database.ManagerConfig
	if err := config.Inject("databases", &dbManagerConfig); err != nil {
		appSlog.Error("Failed to inject database configurations", "error", err)
		os.Exit(1)
	}

	// Create the DatabaseManager by injecting the configuration.
	dbManager, err := database.NewManager(dbManagerConfig, database.WithLogger(appSlog))
	if err != nil {
		appSlog.Error("Failed to create database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// 2. Get DB connection and instantiate the seeder
	provider, err := dbManager.Connection("my_postgres")
	if err != nil {
		appSlog.Error("Failed to get 'my_postgres' connection", "error", err)
		os.Exit(1)
	}

	// Type-assert the provider to get the underlying GORM instance.
	sqlProvider, _ := provider.(database.SQLProvider)
	writerDB := sqlProvider.Gorm().(*gorm.DB)
	seeder := dgseeder.New(writerDB, appSlog)

	// 3. Register seeder functions and set order
	seeder.Register("users", seeders.UserSeeder)
	// seeder.Register("products", seeders.ProductSeeder)
	seeder.SetOrder([]string{"users", "products"})

	// 4. Run the seeders (atomically)
	if err := seeder.RunAllWithTransaction(); err != nil {
		appSlog.Error("Database seeding failed", "error", err)
		os.Exit(1)
	}

	appSlog.Info("✅ Database seeding completed successfully!")
}
```