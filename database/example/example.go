package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/donnigundala/dgcore/config" // Assuming a framework config package
	"github.com/donnigundala/dgcore/database"
	"github.com/donnigundala/dgcore/database/contracts"
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

	log.Println("--- Database Package Example (Framework-Idiomatic) ---")

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
	mgr.SetLogger(appSlog) // Set the application's slog logger on the manager

	for name, cfg := range dbConfigs {
		// Assign the trace ID key from the config (e.g., from YAML) to each config
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
