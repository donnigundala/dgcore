package database

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// HealthChecker periodically pings database connections to ensure they are live.
type HealthChecker struct {
	manager    *DatabaseManager
	interval   time.Duration
	ticker     *time.Ticker
	quit       chan struct{}
	logger     *slog.Logger
	checkNow   chan bool
	endpoints  []string
}

// NewHealthChecker creates a new checker for the given DatabaseManager.
func NewHealthChecker(m *DatabaseManager, interval time.Duration, endpoints []string, logger *slog.Logger) *HealthChecker {
	return &HealthChecker{
		manager:    m,
		interval:   interval,
		quit:       make(chan struct{}),
		logger:     logger.With("component", "db_health_checker"),
		checkNow:   make(chan bool, 1),
		endpoints:  endpoints,
	}
}

// Start begins the periodic health checks.
func (hc *HealthChecker) Start() {
	if hc.interval <= 0 {
		hc.logger.Warn("Health checker interval is zero or negative, checker will not start.")
		return
	}
	hc.ticker = time.NewTicker(hc.interval)
	hc.logger.Info("Starting database health checker...", "interval", hc.interval)
	go func() {
		for {
			select {
			case <-hc.ticker.C:
				hc.checkAll()
			case <-hc.checkNow:
				hc.checkAll()
			case <-hc.quit:
				hc.ticker.Stop()
				return
			}
		}
	}()
}

// Stop terminates the health checker.
func (hc *HealthChecker) Stop() {
	hc.logger.Info("Stopping database health checker...")
	close(hc.quit)
}

// CheckAllNow triggers an immediate health check for all connections.
func (hc *HealthChecker) CheckAllNow() {
	select {
	case hc.checkNow <- true:
	default:
		hc.logger.Debug("Immediate health check already in progress, skipping.")
	}
}

func (hc *HealthChecker) checkAll() {
	hc.logger.Debug("Performing health check on database connections...")

	// Safely get a copy of the connections map from the manager
	hc.manager.mu.RLock()
	connections := make(map[string]Provider, len(hc.manager.connections))
	for name, conn := range hc.manager.connections {
		connections[name] = conn
	}
	hc.manager.mu.RUnlock()

	endpointsToCheck := hc.endpoints
	if len(endpointsToCheck) == 0 {
		for name := range connections {
			endpointsToCheck = append(endpointsToCheck, name)
		}
	}

	var wg sync.WaitGroup
	for _, name := range endpointsToCheck {
		conn, ok := connections[name]
		if !ok {
			hc.logger.Warn("Cannot perform health check on unknown connection", "connection", name)
			continue
		}

		wg.Add(1)
		go func(n string, c Provider) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := c.Ping(ctx); err != nil {
				hc.logger.Error("Database health check failed", "connection", n, "error", err)
			} else {
				hc.logger.Debug("Database health check successful", "connection", n)
			}
		}(name, conn)
	}
	wg.Wait()
}
