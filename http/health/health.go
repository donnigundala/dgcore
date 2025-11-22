package health

import (
	"context"
	"sync"
	"time"
)

// Checker defines the interface for health checks.
type Checker interface {
	Check(ctx context.Context) error
	Name() string
}

// Manager manages health checks.
type Manager struct {
	checks  []Checker
	mu      sync.RWMutex
	timeout time.Duration
}

// NewManager creates a new health check manager.
func NewManager() *Manager {
	return &Manager{
		checks:  make([]Checker, 0),
		timeout: 5 * time.Second, // Default 5 seconds
	}
}

// AddCheck adds a health check to the manager.
func (m *Manager) AddCheck(checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checks = append(m.checks, checker)
}

// SetTimeout sets the timeout for health checks.
func (m *Manager) SetTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeout = timeout
}

// CheckAll runs all registered health checks.
func (m *Manager) CheckAll(ctx context.Context) map[string]error {
	m.mu.RLock()
	checks := make([]Checker, len(m.checks))
	copy(checks, m.checks)
	timeout := m.timeout
	m.mu.RUnlock()

	results := make(map[string]error)
	resultsMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, checker := range checks {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			err := c.Check(ctx)
			resultsMu.Lock()
			results[c.Name()] = err
			resultsMu.Unlock()
		}(checker)
	}

	wg.Wait()
	return results
}

// IsHealthy returns true if all health checks pass.
func (m *Manager) IsHealthy(ctx context.Context) bool {
	results := m.CheckAll(ctx)
	for _, err := range results {
		if err != nil {
			return false
		}
	}
	return true
}
