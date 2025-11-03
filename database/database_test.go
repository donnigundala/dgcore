package database_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/donnigundala/dgcore/database"
	"github.com/donnigundala/dgcore/database/config"
	"github.com/donnigundala/dgcore/database/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestContextKey is a custom type for context keys to avoid collisions in tests.
type TestContextKey string

const (
	TestTraceIDContextKey TestContextKey = "test-trace-id"
)

// MockMetricsProvider is a simple mock for testing metrics.
type MockMetricsProvider struct {
	mu          sync.Mutex
	IncCalls    map[string]int
	GaugeValues map[string]float64
}

func NewMockMetricsProvider() *MockMetricsProvider {
	return &MockMetricsProvider{
		IncCalls:    make(map[string]int),
		GaugeValues: make(map[string]float64),
	}
}

func (m *MockMetricsProvider) Inc(name string, labels ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name + "_" + fmt.Sprintf("%v", labels)
	m.IncCalls[key]++
}

func (m *MockMetricsProvider) SetGauge(name string, value float64, labels ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name + "_" + fmt.Sprintf("%v", labels)
	m.GaugeValues[key] = value
}

func (m *MockMetricsProvider) Observe(name string, value float64, labels ...string) {
	// Not implemented for this mock
}

func TestManagerSingleton(t *testing.T) {
	mgr1 := database.Manager()
	mgr2 := database.Manager()
	assert.Same(t, mgr1, mgr2, "Manager() should return a singleton instance")
}

func TestManagerRegisterAndGet(t *testing.T) {
	// Use the non-singleton manager for isolated testing
	mgr := database.NewManager()

	cfg := &contracts.Config{
		Driver: contracts.ProviderSQL,
		SQL: &config.SQLConfig{
			DriverName: "sqlite",
			Primary:    &config.SQLConnectionDetails{DBName: config.Secret{Value: ":memory:"}},
		},
	}

	mgr.Register("test_db", cfg)

	// Test public behavior: Get should return nil before ConnectAll
	provider := mgr.Get("test_db")
	assert.Nil(t, provider, "Provider should be nil before ConnectAll")

	// Test getting a non-existent provider
	nonExistentProvider := mgr.Get("non_existent_db")
	assert.Nil(t, nonExistentProvider, "Getting a non-existent provider should return nil")
}

func TestConnectAllAndSQLite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a discard logger for tests to avoid polluting stdout
	discardLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a new manager instance for a clean test state
	mgr := database.NewManager()
	mgr.SetLogger(discardLogger)

	metrics := NewMockMetricsProvider()

	cfg := &contracts.Config{
		Driver: contracts.ProviderSQL,
		SQL: &config.SQLConfig{
			DriverName: "sqlite",
			Primary:    &config.SQLConnectionDetails{DBName: config.Secret{Value: "file::memory:?cache=shared"}},
			LogLevel:   config.LogLevelInfo,
		},
		Metrics:    metrics,
		TraceIDKey: string(TestTraceIDContextKey),
	}

	mgr.Register("sqlite_test_db", cfg)

	connCtx := context.WithValue(ctx, TestTraceIDContextKey, "test-conn-trace")
	err := mgr.ConnectAll(connCtx)
	require.NoError(t, err, "ConnectAll should not return an error")

	provider := mgr.Get("sqlite_test_db")
	require.NotNil(t, provider, "Provider should not be nil after ConnectAll")

	// Test Ping
	err = provider.Ping(connCtx)
	assert.NoError(t, err, "Ping should succeed")

	// Test GetWriter and basic GORM operation
	writerDB := provider.GetWriter().(*gorm.DB)
	require.NotNil(t, writerDB, "Writer DB should not be nil")

	// AutoMigrate a dummy table
	err = writerDB.WithContext(connCtx).AutoMigrate(&TestModel{})
	require.NoError(t, err, "AutoMigrate should succeed")

	// Create a record
	newRecord := TestModel{Name: "Test Record"}
	err = writerDB.WithContext(connCtx).Create(&newRecord).Error
	require.NoError(t, err, "Create should succeed")
	assert.Greater(t, newRecord.ID, uint(0), "ID should be assigned")

	// Read the record
	readRecord := TestModel{}
	readerDB := provider.GetReader().(*gorm.DB)
	require.NotNil(t, readerDB, "Reader DB should not be nil")

	err = readerDB.WithContext(connCtx).First(&readRecord, newRecord.ID).Error
	require.NoError(t, err, "Read should succeed")
	assert.Equal(t, newRecord.Name, readRecord.Name, "Read record name should match")

	// Test metrics
	assert.Equal(t, float64(1), metrics.GaugeValues["db_primary_healthy_[primary]"], "Primary healthy gauge should be 1")

	// Test Close
	err = mgr.Close()
	assert.NoError(t, err, "Close should succeed")
}

type TestModel struct {
	ID   uint `gorm:"primarykey"`
	Name string
}

// TODO: Add more comprehensive tests for:
// - PostgreSQL (requires Docker or real instance)
// - MySQL (requires Docker or real instance)
// - MongoDB (requires Docker or real instance/mocking library)
// - Failover scenarios
// - Read/Write splitting with replicas
// - Connection pool settings
// - TLS configuration
// - Error handling
// - Metrics reporting for all events
// - Slog context awareness
