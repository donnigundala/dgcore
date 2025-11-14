package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Runnable defines a common interface for servers that can be started and stopped.
type Runnable interface {
	Start() error
	Shutdown(ctx context.Context) error
}

// registry wraps a Runnable with metadata.
type registry struct {
	Name    string
	Enabled bool
	Runner  Runnable
}

// Manager controls the lifecycle of multiple servers.
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
		servers:         make(map[string]*registry),
		shutdownTimeout: 15 * time.Second, // A sensible default shutdown timeout.
	}

	for _, opt := range opts {
		opt(m)
	}

	// If no logger was provided, create a default one.
	if m.logger == nil {
		m.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// Create a sub-logger that automatically adds the "component" field.
	m.logger = m.logger.With("component", "server-manager")
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

// Register adds a named server to the manager.
func (m *Manager) Register(name string, server Runnable) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[name]; exists {
		m.logger.Warn("server with the same name already registered", "name", name)
		return
	}

	m.servers[name] = &registry{
		Name:    name,
		Enabled: true,
		Runner:  server,
	}
}

// Enable marks a server as active.
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
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
	m.mu.Lock()
	defer m.mu.Unlock()

	srv, ok := m.servers[name]
	if !ok {
		return fmt.Errorf("server %q not found", name)
	}
	srv.Enabled = false
	return nil
}

// RunAll starts all enabled servers concurrently and blocks until the context is canceled or a server fails.
func (m *Manager) RunAll(ctx context.Context) error {
	return m.runServers(ctx)
}

// Run starts specific servers by name.
func (m *Manager) Run(ctx context.Context, names ...string) error {
	return m.runServers(ctx, names...)
}

// runServers is the internal orchestration logic.
func (m *Manager) runServers(ctx context.Context, names ...string) error {
	targets, targetNames := m.selectServers(names...)
	if len(targets) == 0 {
		return errors.New("no servers selected to run")
	}

	m.logger.Info("starting servers", "servers", strings.Join(targetNames, ", "))

	wg := &sync.WaitGroup{}
	errCh := make(chan error, len(targets))

	// Start all target servers in separate goroutines.
	for _, srv := range targets {
		wg.Add(1)
		go func(s *registry) {
			defer wg.Done()
			m.logger.Info("server started", "name", s.Name)
			if err := s.Runner.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("server %q failed: %w", s.Name, err)
			}
		}(srv)
	}

	var runErr error
	// Wait for a shutdown signal (context cancellation) or a server error.
	select {
	case <-ctx.Done():
		m.logger.Info("shutdown signal received, initiating graceful shutdown")
		runErr = ctx.Err() // Capture context cancellation error (e.g., Canceled, DeadlineExceeded)
	case err := <-errCh:
		m.logger.Error("a server failed, initiating shutdown", "error", err)
		runErr = err // Capture the first server error
	}

	// Create a context for the shutdown process with the configured timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), m.shutdownTimeout)
	defer cancel()

	// Initiate shutdown for all running servers.
	m.logger.Info("shutting down servers", "timeout", m.shutdownTimeout.String())
	shutdownErrCh := make(chan error, len(targets))
	for _, srv := range targets {
		go func(s *registry) {
			if err := s.Runner.Shutdown(shutdownCtx); err != nil {
				shutdownErrCh <- fmt.Errorf("shutdown failed for server %q: %w", s.Name, err)
			}
		}(srv)
	}

	// Wait for all server-starting goroutines to finish.
	wg.Wait()
	close(errCh)
	close(shutdownErrCh)

	// Collect all shutdown errors.
	var shutdownErrs []error
	for err := range shutdownErrCh {
		m.logger.Error("error during server shutdown", "error", err)
		shutdownErrs = append(shutdownErrs, err)
	}

	// Return the initial error that triggered the shutdown, combined with any shutdown errors.
	return errors.Join(runErr, errors.Join(shutdownErrs...))
}

// selectServers filters servers based on the provided names.
// If no names are provided, it returns all enabled servers.
func (m *Manager) selectServers(names ...string) ([]*registry, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var targets []*registry
	var targetNames []string

	// If no names are provided, select all enabled servers.
	if len(names) == 0 {
		for _, srv := range m.servers {
			if srv.Enabled {
				targets = append(targets, srv)
				targetNames = append(targetNames, srv.Name)
			}
		}
		return targets, targetNames
	}

	// Otherwise, select specific servers by name.
	for _, n := range names {
		if srv, ok := m.servers[n]; ok && srv.Enabled {
			targets = append(targets, srv)
			targetNames = append(targetNames, srv.Name)
		} else {
			m.logger.Warn("server not found or disabled, skipping", "name", n)
		}
	}

	return targets, targetNames
}
