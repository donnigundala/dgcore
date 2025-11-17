package broker

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// Manager handles the lifecycle of all broker connections.
type Manager struct {
	mu      sync.RWMutex
	brokers map[string]Provider
	logger  *slog.Logger
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithLogger provides a slog logger for the broker manager.
func WithLogger(logger *slog.Logger) ManagerOption {
	return func(m *Manager) {
		if logger != nil {
			m.logger = logger.With("component", "broker_manager")
		}
	}
}

// NewManager creates a new instance of the Manager from a map of configurations.
func NewManager(configs map[string]*Config, opts ...ManagerOption) (*Manager, error) {
	m := &Manager{
		brokers: make(map[string]Provider),
		logger:  slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "broker_manager"),
	}

	for _, opt := range opts {
		opt(m)
	}

	if len(configs) == 0 {
		m.logger.Warn("No broker configurations provided.")
		return m, nil // Return an empty manager
	}

	m.logger.Info("Initializing broker manager...", "config_count", len(configs))

	for name, cfg := range configs {
		provider, err := newProvider(cfg, m.logger)
		if err != nil {
			m.logger.Error("Failed to connect to broker.", "broker_name", name, "error", err)
			return nil, fmt.Errorf("failed to connect to broker '%s': %w", name, err)
		}
		m.brokers[name] = provider
		m.logger.Info("Successfully connected to broker.", "broker_name", name, "driver", cfg.Driver)
	}

	return m, nil
}

// Broker returns the broker provider for the given name.
// It returns an error if the provider is not found.
func (m *Manager) Broker(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.brokers[name]
	if !ok {
		return nil, fmt.Errorf("broker provider '%s' not configured", name)
	}
	return provider, nil
}

// MustBroker is like Broker but panics if the provider is not found.
func (m *Manager) MustBroker(name string) Provider {
	provider, err := m.Broker(name)
	if err != nil {
		panic(err)
	}
	return provider
}

// Close gracefully closes all managed broker connections.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Closing all broker connections...")
	var allErrors []error
	for name, conn := range m.brokers {
		if err := conn.Close(); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed closing broker '%s': %w", name, err))
		}
		delete(m.brokers, name)
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

// newProvider acts as an internal factory for creating a broker Provider.
func newProvider(cfg *Config, logger *slog.Logger) (Provider, error) {
	connLogger := logger.With("driver", cfg.Driver)
	switch cfg.Driver {
	case "kafka":
		return newKafkaProvider(cfg, connLogger)
	case "rabbitmq":
		return newRabbitMQProvider(cfg, connLogger)
	case "nats":
		return newNATSProvider(cfg, connLogger)
	default:
		return nil, fmt.Errorf("unsupported broker driver: %s", cfg.Driver)
	}
}
