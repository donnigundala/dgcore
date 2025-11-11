package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// Manager handles the lifecycle of all database connections.
type Manager struct {
	mu                sync.RWMutex
	connections       map[string]Provider
	defaultConnection string
	logger            *slog.Logger
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithLogger provides a slog logger for the database manager.
func WithLogger(logger *slog.Logger) ManagerOption {
	return func(m *Manager) {
		if logger != nil {
			m.logger = logger.With("component", "database_manager")
		}
	}
}

// NewManager creates a new instance of the Manager from a configuration struct.
func NewManager(cfg Config, opts ...ManagerOption) (*Manager, error) {
	m := &Manager{
		connections: make(map[string]Provider),
		logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "database_manager"),
	}

	for _, opt := range opts {
		opt(m)
	}

	if len(cfg.Connections) == 0 {
		m.logger.Warn("No database connections configured.")
		return m, nil
	}

	m.logger.Info("Initializing database manager...", "connection_count", len(cfg.Connections))

	for name, connCfg := range cfg.Connections {
		provider, err := newProvider(context.Background(), name, connCfg, m.logger)
		if err != nil {
			m.logger.Error("Failed to connect to database", "connection", name, "error", err)
			return nil, fmt.Errorf("failed to connect to '%s': %w", name, err)
		}
		m.connections[name] = provider
		m.logger.Info("Successfully connected to database.", "connection", name, "driver", connCfg.Driver)
	}

	m.defaultConnection = cfg.DefaultConnection
	if _, ok := m.connections[m.defaultConnection]; m.defaultConnection != "" && !ok {
		return nil, fmt.Errorf("default connection '%s' not found in configured connections", m.defaultConnection)
	}

	return m, nil
}

// Connection returns a specific database provider by name.
// If the name is empty, it returns the default connection.
func (m *Manager) Connection(name ...string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var connName string
	if len(name) > 0 && name[0] != "" {
		connName = name[0]
	} else {
		connName = m.defaultConnection
	}

	if connName == "" {
		return nil, errors.New("no database connection name specified and no default connection is set")
	}

	provider, ok := m.connections[connName]
	if !ok {
		return nil, fmt.Errorf("database connection '%s' not configured", connName)
	}
	return provider, nil
}

// MustConnection is like Connection but panics if the provider is not found.
func (m *Manager) MustConnection(name ...string) Provider {
	provider, err := m.Connection(name...)
	if err != nil {
		panic(err)
	}
	return provider
}

// Close gracefully closes all managed database connections.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Closing all database connections...")
	var allErrors []error
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			m.logger.Error("Failed to close connection for database.", "connection", name, "error", err)
			allErrors = append(allErrors, fmt.Errorf("failed closing '%s': %w", name, err))
		}
		delete(m.connections, name)
	}

	if len(allErrors) > 0 {
		m.logger.Error("Finished closing connections with errors.", "error_count", len(allErrors))
		return errors.Join(allErrors...)
	}

	m.logger.Info("All connections closed successfully.")
	return nil
}
