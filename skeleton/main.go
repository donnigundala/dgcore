package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/errors"
	"github.com/donnigundala/dgcore/foundation"
	coreHTTP "github.com/donnigundala/dgcore/http"
	"github.com/donnigundala/dgcore/logging"

	"example-app/routes"
)

// AppConfig represents the application configuration
type AppConfig struct {
	Name  string `mapstructure:"name"`
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
	Port  int    `mapstructure:"port"`
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

	// Load configuration files
	config.Load()

	// Inject app configuration
	var appConfig AppConfig
	if err := config.Inject("app", &appConfig); err != nil {
		logger.Error("Failed to load app configuration", "error", err)
		os.Exit(1)
	}

	// Use debug level if debug mode is enabled
	if appConfig.Debug {
		logger = logging.New(logging.Config{
			Level:      slog.LevelDebug,
			Output:     os.Stdout,
			JSONFormat: false,
			AddSource:  true,
		})
		logging.SetDefault(logger)
	}

	logger.Info("Starting application",
		"name", appConfig.Name,
		"env", appConfig.Env,
		"debug", appConfig.Debug,
	)

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

	// Get Kernel and create HTTP server
	kernelInstance, err := app.Make("kernel")
	if err != nil {
		wrappedErr := errors.Wrap(err, "failed to resolve kernel")
		logger.Error("Failed to resolve kernel", "error", wrappedErr)
		os.Exit(1)
	}
	kernel := kernelInstance.(http.Kernel)

	// Start HTTP server with port from config
	addr := fmt.Sprintf(":%d", appConfig.Port)
	cfg := coreHTTP.Config{Addr: addr}
	server := coreHTTP.NewHTTPServer(cfg, kernel)

	logger.Info("Starting HTTP server", "addr", cfg.Addr)
	if err := server.Start(); err != nil {
		wrappedErr := errors.Wrap(err, "server failed to start").
			WithCode("SERVER_START_FAILED")
		logger.Error("Server failed", "error", wrappedErr)
		os.Exit(1)
	}
}
