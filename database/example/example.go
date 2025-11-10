package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/database"
)

// This example assumes you have a `config` directory next to your executable
// with a `database.yaml` file inside it.

func main() {
	// 1. Initialize a logger for the application.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 2. Load all configurations from the default path ("./config").
	config.Load()

	// 3. Define a struct to hold the database manager configuration.
	// The top-level key in your YAML file should be "databases".
	var dbManagerConfig database.ManagerConfig
	if err := config.Inject("databases", &dbManagerConfig); err != nil {
		slog.Error("Failed to inject database configuration", "error", err)
		os.Exit(1)
	}

	// 4. Create a new DatabaseManager by injecting the loaded configuration.
	dbManager, err := database.NewManager(dbManagerConfig, database.WithLogger(logger))
	if err != nil {
		slog.Error("Failed to create database manager", "error", err)
		os.Exit(1)
	}
	defer func(dbManager *database.DatabaseManager) {
		err := dbManager.Close()
		if err != nil {
			slog.Error("Failed to close database manager", "error", err)
		}
	}(dbManager)

	slog.Info("Database manager initialized successfully.")

	// 5. Get the default database connection.
	// Use MustConnection for convenience; it panics if not found.
	// Use Connection() for safe error handling.
	defaultDB, err := dbManager.Connection() // Gets the default connection
	if err != nil {
		slog.Error("Failed to get default database connection", "error", err)
		os.Exit(1)
	}

	// 6. Ping the database to check the connection.
	slog.Info("Pinging default database...")
	if err := defaultDB.Ping(context.Background()); err != nil {
		slog.Error("Failed to ping default database", "error", err)
	} else {
		slog.Info("Default database ping successful!")
	}

	// You can also get a specific connection by name
	// myPostgres, err := dbManager.Connection("my_postgres")
	// ...
}
