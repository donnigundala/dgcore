# `core/firebase` Package

## Overview

The `firebase` package provides a robust and configurable way to integrate Firebase services into your application. It follows the framework's Dependency Injection (DI) pattern, allowing you to manage multiple Firebase project connections cleanly and efficiently.

## Features

- **Manager Pattern**: A central `FirebaseManager` handles the lifecycle of multiple named Firebase app instances, consistent with other framework packages like `database` and `cache`.
- **Framework-Integrated Configuration**: Leverages the central `config` package, allowing configuration via YAML files and environment variables.
- **Secure Credential Handling**: Designed to load credentials from environment variables, preventing sensitive data from being hardcoded in your repository.
- **Structured Logging**: Integrates with `slog` for consistent, structured logging.

## Configuration

The Firebase manager is configured through a `firebase` block in your application's `config.yaml`. You can define multiple named app connections.

**It is strongly recommended to provide credentials via environment variables for security.**

### YAML Configuration

```yaml
# your-app/config/firebase.yaml
firebase:
  # A 'default' app connection
  default:
    # For local development, you can use a file path.
    credentials_file: "./config/firebase-credentials.json"
    # In production, it's better to use an environment variable for the JSON content.
    credentials_json: ${FIREBASE_CREDENTIALS_JSON}

  # A second app connection for a different project
  another_project:
    credentials_json: ${ANOTHER_PROJECT_CREDENTIALS_JSON}
```

### Environment Variable Configuration

You can provide the entire credentials JSON content directly as an environment variable.

```sh
# For the 'default' app
export FIREBASE_DEFAULT_CREDENTIALS_JSON='{"type": "service_account", "project_id": "...", ...}'

# For the 'another_project' app
export FIREBASE_ANOTHER_PROJECT_CREDENTIALS_JSON='{"type": "service_account", "project_id": "...", ...}'
```

## Usage

The idiomatic workflow involves loading configurations into a map and injecting it into the `FirebaseManager`.

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/firebase"
)

func main() {
	// 1. Bootstrap logger and config
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, nil))
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

	// 4. Get a specific Firebase app instance
	defaultApp := fbManager.MustApp("default")

	// 5. Get the corresponding config to pass to service clients
	defaultConfig := firebaseConfigs["default"]

	// 6. Use the app to get a service client, e.g., FCM
	fcmClient, err := firebase.NewFCMClient(defaultApp, defaultConfig, appSlog)
	if err != nil {
		slog.Error("Failed to get Firebase FCM client", "error", err)
		os.Exit(1)
	}

	// You can also get other clients like Auth
	// authClient, err := defaultApp.Auth(ctx)
	// ...

	slog.Info("Successfully initialized Firebase and got FCM client.")

	// Now you can use fcmClient to send notifications
}
```