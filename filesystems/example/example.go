package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/filesystems"
	"github.com/google/uuid"
)

// In a real application, this would be defined in a shared internal package.
type CtxKey string

const TraceIDKey = CtxKey("trace_id")

func main() {
	// 1. Create a logger.
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
	fs, err := filesystems.NewManager(fsManagerConfig,
		filesystems.WithLogger(logger),
		filesystems.WithTraceIDKey(TraceIDKey),
	)
	if err != nil {
		logger.Error("Failed to create filesystem manager", "error", err)
		os.Exit(1)
	}

	logger.Info("=== Filesystem Manager Example Starting ===")

	// 5. Create a context with a trace ID.
	traceID := uuid.New().String()
	ctx := context.WithValue(context.Background(), TraceIDKey, traceID)

	logger.InfoContext(ctx, "Starting file operations with trace ID")

	// 6. Use the manager. The trace ID will now be automatically included in all logs.
	localKey := "example-file.txt"
	localData := strings.NewReader("This file will have a trace ID in its logs.")
	err = fs.Upload(ctx, localKey, localData, int64(localData.Len()), filesystems.VisibilityPublic)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to upload file", "key", localKey, "error", err)
	} else {
		logger.InfoContext(ctx, "Successfully uploaded file", "key", localKey)
		fs.Delete(ctx, localKey)
	}
}
