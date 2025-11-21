package main

import (
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/contracts/console"
	"github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/foundation"

	appConsole "example-app/app/console"
	"example-app/routes"
)

func main() {
	basePath, _ := os.Getwd()
	app := foundation.New(basePath)

	// Bind Console Kernel
	app.Singleton("console.kernel", func() interface{} {
		return appConsole.NewKernel(app)
	})

	// Register Routes
	routerInstance, err := app.Make("router")
	if err != nil {
		slog.Error("Failed to resolve router", "error", err)
		os.Exit(1)
	}
	routes.Register(routerInstance.(http.Router))

	// Run Console Kernel
	kernelInstance, err := app.Make("console.kernel")
	if err != nil {
		slog.Error("Failed to resolve console kernel", "error", err)
		os.Exit(1)
	}

	if err := kernelInstance.(console.Kernel).Handle(); err != nil {
		slog.Error("Command failed", "error", err)
		os.Exit(1)
	}
}
