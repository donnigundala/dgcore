package database

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestModel is a simple GORM model for testing purposes.
type TestModel struct {
	gorm.Model
	Name string
}

// TestNewManager_WithSQLite creates a manager with a single SQLite in-memory database,
// connects to it, and performs basic CRUD operations.
func TestNewManager_WithSQLite(t *testing.T) {
	// 1. Define the configuration manually in code (the power of DI for testing).
	dbManagerConfig := Config{
		DefaultConnection: "my_sqlite",
		Connections: map[string]Connection{
			"my_sqlite": {
				Driver: "sql",
				SQL: &SQLConfig{
					DriverName: "sqlite",
					Primary: NodeConfig{
						// Use in-memory SQLite for fast, isolated tests.
						DBName: "file::memory:?cache=shared",
					},
					LogLevel: "silent", // Keep test logs clean.
				},
			},
		},
	}

	// Use a discard logger for tests unless debugging.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// 2. Create the DatabaseManager by injecting the configuration.
	dbManager, err := NewManager(dbManagerConfig, WithLogger(logger))
	require.NoError(t, err, "NewManager should not fail with valid config")
	require.NotNil(t, dbManager, "NewManager should return a non-nil manager")
	defer dbManager.Close()

	// 3. Get the connection.
	// We ask for the default connection.
	provider, err := dbManager.Connection()
	require.NoError(t, err, "Getting the default connection should not fail")
	require.NotNil(t, provider, "The default provider should not be nil")

	// 4. Check the provider type and get the underlying GORM DB.
	sqlProvider, ok := provider.(SQLProvider)
	require.True(t, ok, "Provider should be a SQLProvider")
	gormDB := sqlProvider.Gorm().(*gorm.DB)
	require.NotNil(t, gormDB, "GORM DB instance should not be nil")

	// 5. Perform database operations.
	ctx := context.Background()

	// AutoMigrate a table.
	err = gormDB.WithContext(ctx).AutoMigrate(&TestModel{})
	require.NoError(t, err, "AutoMigrate should succeed")

	// Create a record.
	newUser := TestModel{Name: "Test User"}
	err = gormDB.WithContext(ctx).Create(&newUser).Error
	require.NoError(t, err, "Create operation should succeed")
	assert.Greater(t, newUser.ID, uint(0), "Record should have a non-zero ID after creation")

	// Read the record back.
	var fetchedUser TestModel
	err = gormDB.WithContext(ctx).First(&fetchedUser, newUser.ID).Error
	require.NoError(t, err, "Read operation should succeed")
	assert.Equal(t, "Test User", fetchedUser.Name, "Fetched record should have the correct name")

	// 6. Test closing the manager.
	err = dbManager.Close()
	assert.NoError(t, err, "Closing the manager should not produce an error")
}

// TestNewManager_EmptyConfig tests that the manager can be created with no connections.
func TestNewManager_EmptyConfig(t *testing.T) {
	emptyConfig := Config{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	dbManager, err := NewManager(emptyConfig, WithLogger(logger))
	require.NoError(t, err, "NewManager should not fail with empty config")
	require.NotNil(t, dbManager, "NewManager should return a non-nil manager")

	// Getting a connection should fail.
	_, err = dbManager.Connection("any")
	assert.Error(t, err, "Getting a connection from an empty manager should fail")
}

// TestNewManager_MissingDefault tests that getting the default connection fails
// when the specified default does not exist.
func TestNewManager_MissingDefault(t *testing.T) {
	configWithBadDefault := Config{
		DefaultConnection: "non_existent_default",
		Connections: map[string]Connection{
			"my_sqlite": {
				Driver: "sql",
				SQL: &SQLConfig{
					DriverName: "sqlite",
					Primary:    NodeConfig{DBName: ":memory:"},
				},
			},
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Manager creation itself should fail if the default is not found.
	_, err := NewManager(configWithBadDefault, WithLogger(logger))
	assert.Error(t, err, "NewManager should fail if the default connection is not configured")
}
