# `firebase` Package

## Overview

The `firebase` package provides a robust and configurable way to integrate Firebase services into your application. It is built on the framework's standard patterns, including a central `Manager`, dependency injection, and context-aware operations for superior traceability and performance.

## Core Concepts

### 1. The `Manager`

The `Manager` is the central component that handles the lifecycle of all named Firebase app instances. It is initialized from your application's configuration and provides a single point of access to all Firebase app objects.

### 2. Context-Aware Operations

This is a critical feature of the package. **All operations that involve a network call to Firebase services MUST be performed with a `context.Context`**.

-   **Performance & Reliability**: Passing a context to Firebase operations allows for graceful cancellation and timeouts. If a user's request is canceled or times out, any pending Firebase operation (like sending a push notification) will also be stopped, saving valuable resources.
-   **End-to-End Tracing**: The service clients (like `FCMClient`) are fully integrated with the `ctxutil` package. When you perform an operation (e.g., `fcmClient.Send(ctx, ...)`), the driver automatically uses `ctxutil.LoggerFromContext(ctx)` to get a logger that includes the `request_id`. This enables you to trace a request from the HTTP server all the way down to the specific Firebase API call.

### 3. Service Clients (e.g., `FCMClient`)

Instead of interacting with the raw `firebase.App` object, the package provides dedicated clients for each Firebase service. These clients are designed to be context-aware and provide a clean interface for all operations.

## Full Usage Example

The following example demonstrates how to configure and use the Firebase manager and the FCM client.

### 1. `config/app.yaml`

Define your Firebase app connections in your application's configuration file.

```yaml
firebase:
  # A 'default' app connection
  default:
    # It is highly recommended to provide credentials via an environment variable.
    # The content of the variable should be the full JSON of the service account key.
    credentials_json: ${FIREBASE_CREDENTIALS_JSON}

  # For local development, you can use a file path as an alternative.
  # local_dev:
  #   credentials_file: "./config/firebase-credentials.json"
```

### 2. `main.go` (or your application's entrypoint)

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/firebase"
)

func main() {
	// 1. Bootstrap logger and load configuration
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := config.Load("config/app.yaml"); err != nil {
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
	// Use a background context for initialization.
	ctx := context.Background()
	fbManager, err := firebase.NewManager(ctx, firebaseConfigs, firebase.WithLogger(logger))
	if err != nil {
		logger.Error("failed to create firebase manager", "error", err)
		os.Exit(1)
	}

	// 4. Get a specific Firebase app instance
	defaultApp, err := fbManager.App("default")
	if err != nil {
		logger.Error("failed to get default firebase app", "error", err)
		os.Exit(1)
	}

	// 5. Get the corresponding config to pass to the service client
	defaultConfig := firebaseConfigs["default"]

	// 6. Create a service client (e.g., FCM) from the app instance
	fcmClient, err := firebase.NewFCMClient(defaultApp, defaultConfig, logger)
	if err != nil {
		logger.Error("failed to create FCM client", "error", err)
		os.Exit(1)
	}

	// --- Performing Operations ---
	// Create a sample context, similar to what the server middleware would do.
	requestCtx := context.Background()
	requestCtx = ctxutil.WithLogger(requestCtx, logger.With("request_id", "xyz-789"))

	// 7. Use the client to perform a context-aware operation
	message := &firebase.FCMMessage{
		Title: "Hello from Framework",
		Body:  "Your request was processed!",
	}
	
	// The token would typically come from your database.
	deviceToken := "a_valid_device_token"

	err = fcmClient.Send(requestCtx, deviceToken, message)
	if err != nil {
		// The log entry from the Send operation inside the FCM client will
		// automatically contain "request_id=xyz-789".
		ctxutil.LoggerFromContext(requestCtx).Error("failed to send FCM message", "error", err)
	} else {
		ctxutil.LoggerFromContext(requestCtx).Info("FCM message sent successfully")
	}
}
```
