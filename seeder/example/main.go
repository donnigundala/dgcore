package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/database"
	dbcontracts "github.com/donnigundala/dgcore/database/contracts"
	dgseeder "github.com/donnigundala/dgcore/seeder"
	"gorm.io/gorm"
)

// --- Define Your Models (e.g., in my-app/internal/models) ---

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"unique"`
}

type Product struct {
	gorm.Model
	Name   string
	UserID uint // Foreign key to User
}

// --- Define Your Seeder Functions (e.g., in my-app/database/seeders) ---

func UserSeeder(db *gorm.DB) error {
	users := []User{
		{Name: "Admin User", Email: "admin@example.com"},
		{Name: "Test User", Email: "test@example.com"},
	}
	// Using FirstOrCreate to prevent duplicates on re-runs
	return db.FirstOrCreate(&users).Error
}

func ProductSeeder(db *gorm.DB) error {
	var adminUser User
	if err := db.Where("email = ?", "admin@example.com").First(&adminUser).Error; err != nil {
		return err // Fails if the user seeder didn't run first
	}

	products := []Product{
		{Name: "Laptop", UserID: adminUser.ID},
		{Name: "Mouse", UserID: adminUser.ID},
	}
	return db.FirstOrCreate(&products).Error
}

func main() {
	// --- 1. Standard Application Bootstrap ---
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(appSlog)

	// Load config from file (config.yaml) and environment variables
	config.Load()

	var dbConfigs map[string]*dbcontracts.Config
	if err := config.Inject("databases", &dbConfigs); err != nil {
		appSlog.Error("Failed to inject database configurations", "error", err)
		os.Exit(1)
	}

	dbManager := database.Manager()
	dbManager.SetLogger(appSlog)
	for name, cfg := range dbConfigs {
		dbManager.Register(name, cfg)
	}

	if err := dbManager.ConnectAll(context.Background()); err != nil {
		appSlog.Error("Failed to connect to databases", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// --- 2. Get the Database Connection for Seeding ---
	// We'll seed the 'my_postgres' database connection defined in config.yaml
	provider := dbManager.Get("my_postgres")
	if provider == nil {
		appSlog.Error("Database provider 'my_postgres' not found")
		os.Exit(1)
	}

	writerDB, ok := provider.GetWriter().(*gorm.DB)
	if !ok || writerDB == nil {
		appSlog.Error("Failed to get a valid GORM writer connection")
		os.Exit(1)
	}

	// --- 3. Instantiate, Register, and Run the Seeder ---
	seeder := dgseeder.New(writerDB, appSlog)

	// Register seeders programmatically
	seeder.Register("users", UserSeeder)
	seeder.Register("products", ProductSeeder)

	// Set a specific run order because products depend on users
	seeder.SetOrder([]string{"users", "products"})

	// Use RunAllWithTransaction to ensure atomicity
	if err := seeder.RunAllWithTransaction(); err != nil {
		appSlog.Error("Database seeding failed", "error", err)
		os.Exit(1)
	}

	appSlog.Info("✅ Database seeding completed successfully!")
}
