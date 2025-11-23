package foundation

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownHook is a function that is called during shutdown.
type ShutdownHook func()

// shutdownManager manages graceful shutdown.
type shutdownManager struct {
	hooks   []ShutdownHook
	mu      sync.Mutex
	timeout time.Duration
}

// newShutdownManager creates a new shutdown manager.
func newShutdownManager() *shutdownManager {
	return &shutdownManager{
		hooks:   make([]ShutdownHook, 0),
		timeout: 30 * time.Second, // Default 30 seconds
	}
}

// RegisterShutdownHook registers a function to be called during shutdown.
func (a *Application) RegisterShutdownHook(hook ShutdownHook) {
	a.shutdown.mu.Lock()
	defer a.shutdown.mu.Unlock()
	a.shutdown.hooks = append(a.shutdown.hooks, hook)
}

// SetShutdownTimeout sets the maximum time to wait for shutdown to complete.
func (a *Application) SetShutdownTimeout(timeout time.Duration) {
	a.shutdown.mu.Lock()
	defer a.shutdown.mu.Unlock()
	a.shutdown.timeout = timeout
}

// Shutdown executes all registered shutdown hooks.
func (a *Application) Shutdown(ctx context.Context) error {
	a.shutdown.mu.Lock()
	hooks := make([]ShutdownHook, len(a.shutdown.hooks))
	copy(hooks, a.shutdown.hooks)
	a.shutdown.mu.Unlock()

	// Execute hooks in reverse order (LIFO)
	for i := len(hooks) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			hooks[i]()
		}
	}

	return nil
}

// WaitForShutdown blocks until a shutdown signal is received.
// It handles SIGTERM and SIGINT signals and executes cleanup hooks.
func (a *Application) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Wait for signal
	sig := <-quit

	// Get logger if available
	var logger interface{}
	if l, err := a.Make("logger"); err == nil {
		logger = l
	}

	if logger != nil {
		// Log shutdown signal (assuming logger has Info method)
		// This is a best-effort log
		type infoLogger interface {
			Info(msg string, args ...interface{})
		}
		if il, ok := logger.(infoLogger); ok {
			il.Info("Received shutdown signal", "signal", sig.String())
		}
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdown.timeout)
	defer cancel()

	// Execute shutdown hooks
	if err := a.Shutdown(ctx); err != nil {
		if logger != nil {
			type errorLogger interface {
				Error(msg string, args ...interface{})
			}
			if el, ok := logger.(errorLogger); ok {
				el.Error("Shutdown error", "error", err)
			}
		}
	}

	if logger != nil {
		type infoLogger interface {
			Info(msg string, args ...interface{})
		}
		if il, ok := logger.(infoLogger); ok {
			il.Info("Application shutdown complete")
		}
	}
}

// GracefulShutdown is a helper that starts a goroutine to wait for shutdown signals.
// It returns a channel that will be closed when shutdown is complete.
func (a *Application) GracefulShutdown() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		a.WaitForShutdown()
		close(done)
	}()
	return done
}
