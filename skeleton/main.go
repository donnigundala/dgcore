package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/foundation"
	coreHTTP "github.com/donnigundala/dgcore/http"

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

	// Load configuration files
	config.Load()

	// Inject app configuration
	var appConfig AppConfig
	if err := config.Inject("app", &appConfig); err != nil {
		slog.Error("Failed to load app configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting application", "name", appConfig.Name, "env", appConfig.Env)

	// Bind the Router
	app.Singleton("router", func() interface{} {
		return coreHTTP.NewRouter()
	})

	// Bind the HTTP Kernel
	app.Singleton("kernel", func() interface{} {
		routerInstance, err := app.Make("router")
		if err != nil {
			panic(err)
		}
		router := routerInstance.(http.Router)
		return coreHTTP.NewKernel(app, router)
	})

	// Register Routes
	routerInstance, err := app.Make("router")
	if err != nil {
		slog.Error("Failed to resolve router", "error", err)
		os.Exit(1)
	}
	routes.Register(routerInstance.(http.Router))

	// Get Kernel and create HTTP server
	kernelInstance, err := app.Make("kernel")
	if err != nil {
		slog.Error("Failed to resolve kernel", "error", err)
		os.Exit(1)
	}
	kernel := kernelInstance.(http.Kernel)

	// Start HTTP server with port from config
	addr := fmt.Sprintf(":%d", appConfig.Port)
	cfg := coreHTTP.Config{Addr: addr}
	server := coreHTTP.NewHTTPServer(cfg, kernel)

	slog.Info("Starting HTTP server", "addr", cfg.Addr)
	if err := server.Start(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
