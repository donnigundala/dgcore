package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// DatabaseManager handles the lifecycle of all database connections.
type DatabaseManager struct {
	mu          sync.RWMutex
	connections map[string]Provider
	configs     map[string]*Config
	baseLogger  *slog.Logger
}

var (
	globalManager *DatabaseManager
	once          sync.Once
)

// Manager returns the global singleton instance of the DatabaseManager.
func Manager() *DatabaseManager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// NewManager creates a new, non-singleton instance of the DatabaseManager.
// This is primarily intended for use in isolated tests.
func NewManager() *DatabaseManager {
	return &DatabaseManager{
		connections: make(map[string]Provider),
		configs:     make(map[string]*Config),
		// Initialize with a default logger, which can be overridden by SetLogger
		baseLogger: slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "database_manager"),
	}
}

// SetLogger allows an external slog.Logger to be injected into the DatabaseManager.
// This logger will be passed down to individual providers.
func (m *DatabaseManager) SetLogger(logger *slog.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if logger != nil {
		m.baseLogger = logger.With("component", "database_manager")
	} else {
		m.baseLogger = slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "database_manager")
	}
	m.baseLogger.Debug("DatabaseManager logger set.")
}

// Register adds a new database configuration to the manager.
func (m *DatabaseManager) Register(name string, cfg *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; exists {
		m.baseLogger.Warn("Configuration for database is being overwritten.", "db_name", name)
	}
	m.configs[name] = cfg
	m.baseLogger.Info("Registered configuration for database.", "db_name", name, "driver", cfg.Driver)
}

// ConnectAll establishes connections for all registered configurations.
func (m *DatabaseManager) ConnectAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.baseLogger.Info("Starting to connect all registered databases...")
	for name, cfg := range m.configs {
		m.baseLogger.Info("Connecting to database.", "db_name", name, "driver", cfg.Driver)
		// Pass the manager's base logger to the New function
		provider, err := New(ctx, cfg, m.baseLogger)
		if err != nil {
			m.baseLogger.Error("Failed to connect to database.", "db_name", name, "error", err)
			return fmt.Errorf("failed to connect to '%s': %w", name, err)
		}
		m.connections[name] = provider
		m.baseLogger.Info("Successfully connected to database.", "db_name", name)
	}
	m.baseLogger.Info("All connections established.")
	return nil
}

// Get returns the database provider for the given name.
func (m *DatabaseManager) Get(name string) Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.connections[name]
	if !ok {
		m.baseLogger.Warn("Attempted to get a non-existent provider.", "db_name", name)
		return nil
	}
	return provider
}

// Close gracefully closes all database connections.
func (m *DatabaseManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.baseLogger.Info("Closing all database connections...")
	var allErrors []error
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			m.baseLogger.Error("Failed to close connection for database.", "db_name", name, "error", err)
			allErrors = append(allErrors, fmt.Errorf("failed closing '%s': %w", name, err))
		}
		delete(m.connections, name)
	}

	if len(allErrors) > 0 {
		m.baseLogger.Error("Finished closing connections with errors.", "error_count", len(allErrors))
		return errors.Join(allErrors...)
	}

	m.baseLogger.Info("All connections closed successfully.")
	return nil
}
