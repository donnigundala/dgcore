package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/firebase"
)

// This example demonstrates how to initialize and use the Firebase package.
// It assumes you have a configuration file (e.g., `config/app.yaml`)
// and have set the FIREBASE_CREDENTIALS_JSON environment variable.

func main() {
	// 1. Bootstrap logger and load configuration
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Load configuration from a specific file.
	if err := config.LoadWithPaths("config/app.yaml"); err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Inject the Firebase configurations into a map
	var firebaseConfigs map[string]*firebase.Config
	if err := config.Inject("firebase", &firebaseConfigs); err != nil {
		logger.Error("failed to inject firebase configs", "error", err)
		os.Exit(1)
	}

	// 3. Initialize the Firebase Manager
	// Use a background context for initialization as it's a one-time setup.
	ctx := context.Background()
	fbManager, err := firebase.NewManager(ctx, firebaseConfigs, firebase.WithLogger(logger))
	if err != nil {
		logger.Error("failed to create firebase manager", "error", err)
		os.Exit(1)
	}
	// The manager doesn't require a Close(), but it's good practice for consistency.
	defer fbManager.Close()

	// 4. Get a specific Firebase app instance
	defaultApp, err := fbManager.App("default")
	if err != nil {
		logger.Error("failed to get default firebase app", "error", err)
		os.Exit(1)
	}

	// 5. Get the corresponding config to pass to the service client
	defaultConfig, ok := firebaseConfigs["default"]
	if !ok {
		logger.Error("configuration for 'default' firebase app not found")
		os.Exit(1)
	}

	// 6. Create a service client (e.g., FCM) from the app instance
	fcmClient, err := firebase.NewFCMClient(defaultApp, defaultConfig, logger)
	if err != nil {
		logger.Error("failed to create FCM client", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ Firebase manager and FCM client initialized successfully.")

	// --- Performing Operations ---
	// Create a sample context, similar to what the server middleware would do.
	requestCtx := context.Background()
	requestCtx = ctxutil.WithLogger(requestCtx, logger.With("request_id", "xyz-123-abc-789"))

	// 7. Use the client to perform a context-aware operation
	message := &firebase.FCMMessage{
		Title: "Hello from Framework",
		Body:  "Your request was processed successfully!",
		Data: map[string]string{
			"orderId": "12345",
		},
	}

	// The device token would typically come from your database.
	// Using a placeholder token will result in an error, which is expected
	// and demonstrates the error logging.
	deviceToken := "a_placeholder_device_token"

	err = fcmClient.Send(requestCtx, deviceToken, message)
	if err != nil {
		// The log entry from the Send operation inside the FCM client will
		// automatically contain "request_id=xyz-123-abc-789".
		ctxutil.LoggerFromContext(requestCtx).Error("failed to send FCM message (as expected with a placeholder token)", "error", err)
	} else {
		ctxutil.LoggerFromContext(requestCtx).Info("FCM message sent successfully")
	}
}
