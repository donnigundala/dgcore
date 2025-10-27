package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog" // Add this import
	"net/http"
	"os" // Add this import for the default logger
	"sync"
	"time"
)

// Runnable defines a common interface for HTTP/gRPC servers.
type Runnable interface {
	Start() error
	Close() error
	Addr() string
	Shutdown(ctx context.Context) error
}

// registry wraps a Runnable with metadata.
type registry struct {
	Name    string
	Enabled bool
	Runner  Runnable
}

// Manager controls lifecycle of multiple servers.
type Manager struct {
	mu              sync.RWMutex
	servers         map[string]*registry
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

// ManagerOption defines an option for configuring a Manager.
type ManagerOption func(*Manager)

// NewManager creates a new server manager.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		servers: make(map[string]*registry),
	}

	for _, opt := range opts {
		opt(m)
	}

	// If no logger was provided, set a sensible default.
	if m.logger == nil {
		m.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// Create a sub-logger that automatically adds the "component" field.
	m.logger = m.logger.With("component", "server manager")
	return m
}

// WithLogger provides a logger for the manager.
func WithLogger(l *slog.Logger) ManagerOption {
	return func(m *Manager) {
		m.logger = l
	}
}

// WithShutdownTimeout sets the timeout for graceful server shutdown.
func WithShutdownTimeout(d time.Duration) ManagerOption {
	return func(m *Manager) {
		m.shutdownTimeout = d
	}
}

// Register adds a named server.
func (m *Manager) Register(name string, server Runnable) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.servers[name] = &registry{
		Name:    name,
		Enabled: true,
		Runner:  server,
	}
}

// Enable marks a server as active.
func (m *Manager) Enable(name string) error {
	m.mu.Lock() // Use a write lock as you're modifying a field
	defer m.mu.Unlock()

	srv, ok := m.servers[name]
	if !ok {
		return fmt.Errorf("server %q not found", name)
	}
	srv.Enabled = true
	return nil
}

// Disable marks a server as inactive.
func (m *Manager) Disable(name string) error {
	m.mu.Lock() // Use a write lock as you're modifying a field
	defer m.mu.Unlock()
	srv, ok := m.servers[name]
	if !ok {
		return fmt.Errorf("server %q not found", name)
	}
	srv.Enabled = false
	return nil
}

// RunAll starts all enabled servers concurrently and blocks until the context is canceled or an error occurs.
func (m *Manager) RunAll(ctx context.Context) error {
	var servers []string
	for name, srv := range m.servers {
		if srv.Enabled {
			servers = append(servers, name)
		}
	}

	err := m.runServers(ctx, servers...)
	if err != nil {
		return fmt.Errorf("error running servers: %w", err)
	}
	return nil
}

// Run runs specific servers by name.
func (m *Manager) Run(ctx context.Context, names ...string) error {
	err := m.runServers(ctx, names...)
	if err != nil {
		return fmt.Errorf("error running servers: %w", err)
	}
	return nil
}

// runServers is the internal orchestration logic.
func (m *Manager) runServers(ctx context.Context, names ...string) error {
	targets := m.selectServers(names...)
	if len(targets) == 0 {
		return fmt.Errorf("no servers selected to run")
	}

	wg := &sync.WaitGroup{}
	errCh := make(chan error, len(targets))

	for _, srv := range targets {
		wg.Add(1)
		go func(s *registry) {
			defer wg.Done()
			m.logger.Info("Starting server", "name", s.Name)
			if err := s.Runner.Start(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					errCh <- fmt.Errorf("%s: %w", s.Name, err)
				}
			}
		}(srv)
	}

	var firstErr error

	// Wait for cancellation or error
	select {
	case <-ctx.Done():
		m.logger.Info("Shutdown signal received, initiating shutdown...")
	case err := <-errCh:
		firstErr = err
		m.logger.Error("Server error", "error", err)
	}

	// Use the configured timeout, with a fallback.
	timeout := m.shutdownTimeout
	if timeout == 0 {
		timeout = 10 * time.Second // A sensible default
	}

	// Graceful shutdown for all *actually started* servers
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	shutdownWg := &sync.WaitGroup{}
	for _, srv := range targets {
		shutdownWg.Add(1)
		go func(s *registry) {
			defer shutdownWg.Done()
			m.logger.Info("Shutting down server", "name", s.Name)
			if err := s.Runner.Shutdown(shutdownCtx); err != nil {
				// This is a reasonable place to just log shutdown errors
				m.logger.Error("Error during server shutdown", "name", s.Name, "error", err)
			}
		}(srv)
	}
	// Wait for all shutdowns to complete
	shutdownWg.Wait()

	// Wait for all server goroutines to complete. This is crucial to prevent leaks.
	wg.Wait()

	// Return the captured error.
	if firstErr != nil {
		return firstErr
	}

	m.logger.Info("All servers stopped gracefully.")
	return nil
}

// Helper: filter servers
func (m *Manager) selectServers(names ...string) []*registry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// if no name provided, run all enabled servers
	if len(names) == 0 {
		targets := make([]*registry, 0, len(m.servers))
		for _, srv := range m.servers {
			if srv.Enabled {
				targets = append(targets, srv)
			}
		}
		return targets
	}

	// run specific servers by name
	targets := make([]*registry, 0, len(names))
	for _, n := range names {
		if srv, ok := m.servers[n]; ok && srv.Enabled {
			targets = append(targets, srv)
		}
	}

	return targets
}
