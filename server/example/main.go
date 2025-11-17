package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/server"
)

func main() {
	// =========================================================================
	// Global Logger
	// =========================================================================
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// =========================================================================
	// Configuration
	// =========================================================================
	// Load configuration from default paths (e.g., ./, ./config/) into the global config store.
	// It's critical to handle this error and fail fast if configuration is malformed.
	if err := config.Load(); err != nil {
		logger.Error("critical error: failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Inject the 'server.http' section of the config into a struct.
	var serverCfg server.Config
	if err := config.Inject("server.http", &serverCfg); err != nil {
		logger.Error("critical error: failed to inject server configuration", "error", err)
		os.Exit(1)
	}

	// =========================================================================
	// Server Manager
	// =========================================================================
	mgr := server.NewManager(
		server.WithLogger(logger),
		server.WithShutdownTimeout(20*time.Second),
	)

	// =========================================================================
	// HTTP Server & Middleware
	// =========================================================================
	mux := http.NewServeMux()
	mux.Handle("/hello", server.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Get the logger from the context using the ctxutil package.
		log := ctxutil.LoggerFromContext(r.Context())
		log.Info("Handling /hello request")

		someOtherFunction(r.Context())

		fmt.Fprintln(w, "Hello, World!")
		return nil
	}))

	var httpHandler http.Handler = mux
	httpHandler = server.RequestIDMiddleware(httpHandler)

	// Create the HTTP server using the injected configuration struct.
	httpServer := server.NewHTTPServer(
		serverCfg,
		httpHandler,
	)

	mgr.Register("http-public", httpServer)

	// =========================================================================
	// Application Start
	// =========================================================================
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("starting server manager")

	if err := mgr.RunAll(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("server manager failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server manager stopped gracefully")
}

func someOtherFunction(ctx context.Context) {
	// This function also uses the ctxutil helper.
	log := ctxutil.LoggerFromContext(ctx)
	log.Info("doing work in another function")
}
