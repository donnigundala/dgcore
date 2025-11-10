package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/firebase"
)

// This example assumes you have a `config` directory next to your executable
// with a `firebase.yaml` and a credentials file inside it.

func main() {
	// 1. Bootstrap logger and config
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(appSlog)

	// Load all configurations from the default path ("./config").
	// It will pick up `firebase.yaml` and handle env vars.
	config.Load()

	// 2. Inject configurations into a map
	var firebaseConfigs map[string]*firebase.Config
	if err := config.Inject("firebase", &firebaseConfigs); err != nil {
		slog.Error("Failed to inject firebase configs", "error", err)
		os.Exit(1)
	}

	// 3. Create the FirebaseManager
	ctx := context.Background()
	fbManager, err := firebase.NewManager(ctx, firebaseConfigs, firebase.WithManagerLogger(appSlog))
	if err != nil {
		slog.Error("Failed to create firebase manager", "error", err)
		os.Exit(1)
	}
	defer fbManager.Close() // Though it's a no-op, it's good practice

	slog.Info("Firebase manager initialized successfully.")

	// 4. Get a specific Firebase app instance
	defaultApp := fbManager.MustApp("default")

	// 5. Get the corresponding config to pass to service clients
	defaultConfig, ok := firebaseConfigs["default"]
	if !ok || defaultConfig == nil {
		slog.Error("Could not find configuration for 'default' firebase app")
		os.Exit(1)
	}

	// 6. Use the app to get a service client, e.g., FCM
	_, err = firebase.NewFCMClient(defaultApp, defaultConfig, appSlog)
	if err != nil {
		slog.Error("Failed to get Firebase FCM client", "error", err)
		os.Exit(1)
	}

	slog.Info("✅ Successfully initialized Firebase and got FCM client.")
}