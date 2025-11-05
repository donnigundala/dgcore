package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/donnigundala/dgcore/config"
	dghttp "github.com/donnigundala/dgcore/server/http"
)

func main() {
	// --- 1. Initialize Application Logger ---
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(appSlog)

	// --- 2. Load Configuration ---
	// This loads config.yaml and merges environment variables.
	// The server's default config is registered via the init() in server/http/config.go
	config.Load()

	// --- 3. Inject Server Configuration into a Struct ---
	var serverCfg dghttp.Config
	if err := config.Inject("server.http", &serverCfg); err != nil {
		appSlog.Error("Failed to inject server configuration", "error", err)
		os.Exit(1)
	}

	// --- 4. Define HTTP Handlers ---
	// In a real application, you would use a router like chi or gin.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		appSlog.Info("Received request", "path", r.URL.Path)
		_, err := fmt.Fprintln(w, "Hello from dg-framework!")
		if err != nil {
			return
		}
	})

	// --- 5. Create and Start the Server ---
	// Create the server from the injected configuration struct.
	// We also inject the application's logger.
	srv := dghttp.NewServerFromConfig(&serverCfg, mux, dghttp.WithLogger(appSlog))

	// Start the server in a non-blocking way.
	err := srv.Start()
	if err != nil {
		return
	}

	// --- 6. Wait for a Shutdown Signal ---
	// This is a blocking call that waits for Ctrl+C or a SIGTERM signal.
	// It will then perform a graceful shutdown with a 10-second timeout.
	srv.WaitForShutdown(10 * time.Second)

	appSlog.Info("Application has shut down.")
}
