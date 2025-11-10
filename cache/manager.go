package cache

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// CacheManager handles the lifecycle of all cache connections.
// It is no longer a singleton and must be created with NewManager.
type CacheManager struct {
	mu          sync.RWMutex
	connections map[string]Provider
	logger      *slog.Logger
}

// ManagerOption configures a CacheManager.
type ManagerOption func(*CacheManager)

// WithLogger provides a slog logger for the cache manager.
func WithLogger(logger *slog.Logger) ManagerOption {
	return func(m *CacheManager) {
		if logger != nil {
			m.logger = logger.With("component", "cache_manager")
		}
	}
}

// NewManager creates a new instance of the CacheManager from a map of configurations.
// This is the primary entry point for creating a cache manager.
func NewManager(configs map[string]*Config, opts ...ManagerOption) (*CacheManager, error) {
	m := &CacheManager{
		connections: make(map[string]Provider),
		// Default to a silent logger, can be overridden by WithLogger.
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "cache_manager"),
	}

	for _, opt := range opts {
		opt(m)
	}

	if len(configs) == 0 {
		m.logger.Warn("No cache configurations provided.")
		return m, nil // Return an empty manager, not an error
	}

	m.logger.Info("Initializing cache manager...", "config_count", len(configs))

	for name, cfg := range configs {
		// Pass the manager's logger down to the provider factory.
		// The factory will then create a sub-logger for the specific driver.
		provider, err := New(cfg, WithProviderLogger(m.logger))
		if err != nil {
			m.logger.Error("Failed to connect to cache.", "cache_name", name, "error", err)
			return nil, fmt.Errorf("failed to connect to cache '%s': %w", name, err)
		}
		m.connections[name] = provider
		m.logger.Info("Successfully connected to cache.", "cache_name", name, "driver", cfg.Driver)
	}

	return m, nil
}

// Get returns the cache provider for the given name.
// It returns an error if the provider is not found.
func (m *CacheManager) Get(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.connections[name]
	if !ok {
		return nil, fmt.Errorf("cache provider '%s' not configured", name)
	}
	return provider, nil
}

// MustGet is like Get but panics if the provider is not found.
func (m *CacheManager) MustGet(name string) Provider {
	provider, err := m.Get(name)
	if err != nil {
		panic(err)
	}
	return provider
}

// Close gracefully closes all managed cache connections.
func (m *CacheManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Closing all cache connections...")
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
