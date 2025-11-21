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
	contractshttp "github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/foundation"
	server "github.com/donnigundala/dgcore/http"
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
	// =========================================================================
	// HTTP Server & Middleware
	// =========================================================================
	router := server.NewRouter()

	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Get the logger from the context using the ctxutil package.
		log := ctxutil.LoggerFromContext(r.Context())
		log.Info("Handling /hello request")

		someOtherFunction(r.Context())

		fmt.Fprintln(w, "Hello, World!")
	})

	// Add a route with parameters
	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User ID requested")
	})

	// Resolve the Kernel from the container
	// In a real app, we would use the Application instance to resolve this.
	// For this example, we'll construct it manually or assume we have the app.
	// Wait, we don't have the 'app' instance here easily accessible as we are in main.
	// But we can create a new router and kernel manually for this example if we want,
	// OR we should really be using the Application to bootstrap everything.

	// Let's switch to using the Application to be proper.
	// We need to import foundation.

	// ... Actually, refactoring the whole main to use foundation.New() might be too big of a change for this step?
	// The plan said "Update example to resolve Kernel and use it as the handler".
	// Let's stick to manual construction for now to keep it simple, or use the router we just created.

	// But Kernel needs an App.
	// Let's create a dummy app or just pass nil if we don't use it in this example?
	// Kernel.Bootstrap() calls app.Boot().

	// Let's do it right. Let's verify if we can easily use foundation.New().
	// We need to import "github.com/donnigundala/dgcore/foundation".

	// For now, let's just use the router directly as before, but wrapped in Kernel if possible?
	// Or better, let's just manually construct the Kernel with a mock app or similar?
	// No, let's update the imports to include foundation and use it.

	// Initialize the application instance
	app := foundation.New("/tmp")

	// Register the router with the application's container
	app.Singleton("router", func() interface{} {
		return router
	})

	// Resolve the HTTP Kernel from the application's container
	kernelInstance, err := app.Make("kernel")
	if err != nil {
		logger.Error("critical error: failed to resolve HTTP Kernel", "error", err)
		os.Exit(1)
	}
	kernel := kernelInstance.(contractshttp.Kernel)

	// Bootstrap the kernel
	kernel.Bootstrap()

	var httpHandler http.Handler = kernel
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
