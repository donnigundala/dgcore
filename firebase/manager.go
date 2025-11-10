package firebase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	fb "firebase.google.com/go/v4"
)

// Manager handles the lifecycle of all Firebase app instances.
type Manager struct {
	mu   sync.RWMutex
	apps map[string]*fb.App
	log  *slog.Logger
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithManagerLogger provides a slog logger for the firebase manager.
func WithManagerLogger(logger *slog.Logger) ManagerOption {
	return func(m *Manager) {
		if logger != nil {
			m.log = logger.With("component", "firebase_manager")
		}
	}
}

// NewManager creates a new instance of the Manager from a map of configurations.
func NewManager(ctx context.Context, configs map[string]*Config, opts ...ManagerOption) (*Manager, error) {
	m := &Manager{
		apps: make(map[string]*fb.App),
		log:  slog.New(slog.NewTextHandler(os.Stdout, nil)).With("component", "firebase_manager"),
	}

	for _, opt := range opts {
		opt(m)
	}

	if len(configs) == 0 {
		m.log.Warn("No Firebase configurations provided.")
		return m, nil // Return an empty manager
	}

	m.log.Info("Initializing Firebase manager...", "config_count", len(configs))

	for name, cfg := range configs {
		// Pass the manager's logger down to the provider factory.
		app, err := New(ctx, cfg, WithLogger(m.log))
		if err != nil {
			m.log.Error("Failed to initialize Firebase app.", "app_name", name, "error", err)
			return nil, fmt.Errorf("failed to initialize Firebase app '%s': %w", name, err)
		}
		m.apps[name] = app.App
		m.log.Info("Successfully initialized Firebase app.", "app_name", name)
	}

	return m, nil
}

// App returns the Firebase app instance for the given name.
// It returns an error if the app is not found.
func (m *Manager) App(name string) (*fb.App, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	app, ok := m.apps[name]
	if !ok {
		return nil, fmt.Errorf("firebase app '%s' not configured or initialized", name)
	}
	return app, nil
}

// MustApp is like App but panics if the instance is not found.
func (m *Manager) MustApp(name string) *fb.App {
	app, err := m.App(name)
	if err != nil {
		panic(err)
	}
	return app
}

// Close is a placeholder for potential future cleanup logic.
// Firebase Go SDK does not require explicit closing of apps.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.log.Info("Closing Firebase manager (no-op).")
	m.apps = make(map[string]*fb.App) // Clear the map
	return nil
}
