package main

import (
	"context" // Import context package
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/database"
	"gorm.io/gorm"
)

// This example assumes you have a `config` directory next to your executable
// with a `database.yaml` file inside it.

// User Example User model for demonstration
type User struct {
	gorm.Model
	Name string
}

func main() {
	// Create a context for the application lifecycle
	ctx := context.Background()

	// 1. Initialize a logger for the application.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 2. Load all configurations from the default path ("./config").
	if err := config.LoadWithPaths("database/example/config/database.yaml"); err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 3. Define a struct to hold the database manager configuration.
	// The top-level key in your YAML file should be "databases".
	var dbManagerConfig database.Config
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
	defer func(dbManager *database.Manager) {
		err := dbManager.Close()
		if err != nil {
			slog.Error("Failed to close database manager", "error", err)
		}
	}(dbManager)

	slog.Info("Database manager initialized successfully.")

	// 5. Get a specific database connection (e.g., my_postgres).
	// Use MustConnection for convenience; it panics if not found.
	// Use Connection() for safe error handling.
	dbProvider, err := dbManager.Connection("my_postgres")
	if err != nil {
		slog.Error("Failed to get 'my_postgres' database connection", "error", err)
		os.Exit(1)
	}

	// 6. Type-assert to get the GORM instance
	sqlProvider, ok := dbProvider.(database.SQLProvider)
	if !ok {
		slog.Error("Connection is not a SQL provider")
		os.Exit(1)
	}
	// Use GormWithContext instead of the deprecated Gorm()
	gormDB := sqlProvider.GormWithContext(ctx)

	// 7. Perform Operations
	slog.Info("--- Performing DB Operations ---")
	if err := gormDB.AutoMigrate(&User{}); err != nil {
		slog.Warn("Could not migrate User model", "error", err)
	}

	// This WRITE operation will automatically go to the PRIMARY node.
	slog.Info("Creating a new user (WRITE operation)...")
	newUser := User{Name: "Framework User"}
	if err := gormDB.Create(&newUser).Error; err != nil {
		slog.Error("Failed to create user", "error", err)
	} else {
		slog.Info("Created user", "name", newUser.Name, "id", newUser.ID)
	}

	// This READ operation will automatically be load-balanced across the REPLICAS.
	slog.Info("Fetching the user (READ operation)...")
	var fetchedUser User
	if err := gormDB.First(&fetchedUser, newUser.ID).Error; err != nil {
		slog.Error("Failed to read user", "error", err)
	} else {
		slog.Info("Fetched user", "name", fetchedUser.Name, "id", fetchedUser.ID)
	}
}
