package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/filesystems"
)

func main() {
	// 1. Create a base application logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// 2. Load configuration from files (e.g., ./config/*.yaml) and environment variables.
	config.Load()

	// 3. Inject the 'filesystems' configuration into a struct.
	var fsManagerConfig filesystems.Config
	if err := config.Inject("filesystems", &fsManagerConfig); err != nil {
		logger.Error("Failed to inject filesystem configuration", "error", err)
		os.Exit(1)
	}

	// 4. Create the FileSystem manager by injecting the config struct.
	fs, err := filesystems.NewManager(fsManagerConfig, filesystems.WithLogger(logger))
	if err != nil {
		logger.Error("Failed to create filesystem manager", "error", err)
		os.Exit(1)
	}

	logger.Info("=== Filesystem Manager Example Starting ===")

	// 5. Simulate a request by creating a context with a request-specific logger.
	// In a real HTTP server, a middleware would do this automatically.
	requestID := ctxutil.NewRequestID()
	requestLogger := logger.With("request_id", requestID)
	ctx := ctxutil.WithLogger(context.Background(), requestLogger)

	// Retrieve the logger from the context to prove it works.
	logFromCtx := ctxutil.LoggerFromContext(ctx)
	logFromCtx.Info("Starting file operations with a context-aware logger.")

	// 6. Use the manager. The manager's internal logs will use the logger provided
	// at initialization, while your application logic can use the context-aware logger.
	localKey := "example-file.txt"
	localData := strings.NewReader("This is some example file content.")
	err = fs.Upload(ctx, localKey, localData, int64(localData.Len()), filesystems.VisibilityPublic)
	if err != nil {
		logFromCtx.Error("Failed to upload file", "key", localKey, "error", err)
	} else {
		logFromCtx.Info("Successfully uploaded file", "key", localKey)
		fs.Delete(ctx, localKey)
	}
}
