package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// CacheManager handles the lifecycle of all cache connections.
type CacheManager struct {
	mu          sync.RWMutex
	connections map[string]Provider
	configs     map[string]*Config
	baseLogger  *slog.Logger
}

var (
	globalManager *CacheManager
	once          sync.Once
)

// Manager returns the global singleton instance of the CacheManager.
func Manager() *CacheManager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// NewManager creates a new, non-singleton instance of the CacheManager.
// This is primarily intended for use in isolated tests.
func NewManager() *CacheManager {
	return &CacheManager{
		connections: make(map[string]Provider),
		configs:     make(map[string]*Config),
		baseLogger:  slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "cache_manager"),
	}
}

// SetLogger allows an external slog.Logger to be injected into the CacheManager.
func (m *CacheManager) SetLogger(logger *slog.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if logger != nil {
		m.baseLogger = logger.With("component", "cache_manager")
	}
}

// Register adds a new cache configuration to the manager.
func (m *CacheManager) Register(name string, cfg *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; exists {
		m.baseLogger.Warn("Configuration for cache is being overwritten.", "cache_name", name)
	}
	m.configs[name] = cfg
	m.baseLogger.Info("Registered configuration for cache.", "cache_name", name, "driver", cfg.Driver)
}

// ConnectAll establishes connections for all registered configurations.
func (m *CacheManager) ConnectAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.baseLogger.Info("Connecting all registered caches...")
	for name, cfg := range m.configs {
		provider, err := New(cfg, WithLogger(m.baseLogger))
		if err != nil {
			m.baseLogger.Error("Failed to connect to cache.", "cache_name", name, "error", err)
			return fmt.Errorf("failed to connect to cache '%s': %w", name, err)
		}
		m.connections[name] = provider
		m.baseLogger.Info("Successfully connected to cache.", "cache_name", name)
	}
	return nil
}

// Get returns the cache provider for the given name.
func (m *CacheManager) Get(name string) Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[name]
}

// Close gracefully closes all cache connections.
func (m *CacheManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.baseLogger.Info("Closing all cache connections...")
	var allErrors []error
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed closing cache '%s': %w", name, err))
		}
		delete(m.connections, name)
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}