package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/errors"
	"github.com/donnigundala/dgcore/foundation"
	coreHTTP "github.com/donnigundala/dgcore/http"
	"github.com/donnigundala/dgcore/http/health"
	"github.com/donnigundala/dgcore/logging"
	"github.com/donnigundala/dgcore/validation"

	"example-app/routes"
)

// AppConfig represents the application configuration with validation.
type AppConfig struct {
	Name  string `mapstructure:"name" validate:"required,min=3"`
	Env   string `mapstructure:"env" validate:"required,oneof=development staging production"`
	Debug bool   `mapstructure:"debug"`
	Port  int    `mapstructure:"port" validate:"required,gte=1,lte=65535"`
}

func main() {
	basePath, _ := os.Getwd()
	app := foundation.New(basePath)

	// Initialize logger
	logLevel := slog.LevelInfo
	logger := logging.New(logging.Config{
		Level:      logLevel,
		Output:     os.Stdout,
		JSONFormat: false,
		AddSource:  false,
	})
	logging.SetDefault(logger)

	// Bind logger to container
	app.Instance("logger", logger)

	// Load configuration files
	config.Load()

	// Inject app configuration
	var appConfig AppConfig
	if err := config.Inject("app", &appConfig); err != nil {
		logger.Error("Failed to load app configuration", "error", err)
		os.Exit(1)
	}

	// Validate configuration
	validator := validation.NewValidator()
	if err := validator.ValidateStruct(context.Background(), &appConfig); err != nil {
		if valErr, ok := err.(*validation.Error); ok {
			logger.Error("Configuration validation failed", "errors", valErr.Errors)
		} else {
			logger.Error("Configuration validation failed", "error", err)
		}
		os.Exit(1)
	}

	logger.Info("Configuration validated successfully")

	// Use debug level if debug mode is enabled
	if appConfig.Debug {
		logger = logging.New(logging.Config{
			Level:      slog.LevelDebug,
			Output:     os.Stdout,
			JSONFormat: false,
			AddSource:  true,
		})
		logging.SetDefault(logger)
		app.Instance("logger", logger)
	}

	logger.Info("Starting application",
		"name", appConfig.Name,
		"env", appConfig.Env,
		"debug", appConfig.Debug,
	)

	// Register shutdown hooks
	app.RegisterShutdownHook(func() {
		logger.Info("Executing shutdown hook: Cleaning up resources...")
		// Add your cleanup logic here:
		// - Close database connections
		// - Close cache connections
		// - Flush logs
		// - etc.
	})

	// Bind the Router
	app.Singleton("router", func() interface{} {
		return coreHTTP.NewRouter()
	})

	// Get router instance
	routerInstance, err := app.Make("router")
	if err != nil {
		wrappedErr := errors.Wrap(err, "failed to resolve router").
			WithCode("ROUTER_RESOLUTION_FAILED").
			WithStatus(500)
		logger.Error("Failed to resolve router", "error", wrappedErr, "code", wrappedErr.Code())
		os.Exit(1)
	}
	router := routerInstance.(http.Router)

	// Setup health checks
	healthManager := health.NewManager()
	healthManager.AddCheck(health.AlwaysHealthy("app"))
	healthManager.AddCheck(health.SimpleCheck("custom", func(ctx context.Context) error {
		// Add your custom health check logic here
		// For example: check database connection, cache connection, etc.
		return nil
	}))

	// Register health check routes
	router.Get("/health/live", health.LivenessHandler())
	router.Get("/health/ready", healthManager.ReadinessHandler())
	router.Get("/health", healthManager.HealthHandler())

	// Apply global middleware
	logger.Info("Applying global middleware")
	router.Use(
		coreHTTP.RecoveryWithDefault(),        // Panic recovery
		coreHTTP.CORSWithDefault(),            // CORS
		coreHTTP.SecurityHeadersWithDefault(), // Security headers
		coreHTTP.BodySizeLimit(10*1024*1024),  // 10MB limit
	)

	// Register Routes
	routes.Register(router)

	logger.Debug("Routes registered successfully")

	// Create HTTP Kernel directly (avoid circular dependency)
	kernel := coreHTTP.NewKernel(app, router)

	// Start HTTP Server
	addr := fmt.Sprintf(":%d", appConfig.Port)
	logger.Info("Starting HTTP server", "port", appConfig.Port)
	server := coreHTTP.NewHTTPServer(coreHTTP.Config{Addr: addr}, kernel)

	// Register server shutdown hook
	app.RegisterShutdownHook(func() {
		logger.Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
		}
	})

	logger.Info("Starting HTTP server", "addr", addr)
	logger.Info("Health endpoints available:")
	logger.Info("  - Liveness:  http://localhost" + addr + "/health/live")
	logger.Info("  - Readiness: http://localhost" + addr + "/health/ready")
	logger.Info("  - Detailed:  http://localhost" + addr + "/health")

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	logger.Info("Application started. Press Ctrl+C to shutdown.")
	app.WaitForShutdown()

	logger.Info("Application stopped gracefully")
}
