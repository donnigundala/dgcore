package database

import (
	"context"
	"fmt"
)

// HealthStatus represents the health of a single database connection.
type HealthStatus struct {
	Name    string `json:"name"`
	Driver  string `json:"driver"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// CheckHealth checks the health of all registered database connections.
func (m *DatabaseManager) CheckHealth(ctx context.Context) []HealthStatus {
	m.mu.RLock()
	defer m.mu.Unlock()

	results := make([]HealthStatus, 0, len(m.connections))
	for name, conn := range m.connections {
		status := HealthStatus{
			Name:   name,
			Driver: string(m.configs[name].Driver),
		}

		if err := conn.Ping(ctx); err != nil {
			status.Healthy = false
			status.Message = "Connection is down."
			status.Error = err.Error()
		} else {
			status.Healthy = true
			status.Message = "Connection is healthy."
		}
		results = append(results, status)
	}
	return results
}

// PrintHealth logs the health status of all connections to the console.
func (m *DatabaseManager) PrintHealth(ctx context.Context) {
	statuses := m.CheckHealth(ctx)
	m.baseLogger.Info("Health Check Report:")

	if len(statuses) == 0 {
		m.baseLogger.Info("No database connections to check.")
		return
	}

	for _, s := range statuses {
		var statusSymbol string
		var logFunc func(msg string, args ...any)
		if s.Healthy {
			statusSymbol = "✔"
			logFunc = m.baseLogger.Info
		} else {
			statusSymbol = "✖"
			logFunc = m.baseLogger.Error
		}

		logFunc(fmt.Sprintf("%s %s (%s): %s", statusSymbol, s.Name, s.Driver, s.Message))
		if !s.Healthy && s.Error != "" {
			logFunc(fmt.Sprintf("  Error: %s", s.Error))
		}
	}
}
