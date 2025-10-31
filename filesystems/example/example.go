package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/donnigundala/dgcore/filesystems"
	"github.com/google/uuid"
)

// In a real application, this would be defined in a shared internal package.
type CtxKey string

const TraceIDKey = CtxKey("trace_id")

func main() {
	// 1. Create a logger.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// 2. Define your storage configuration.
	config := filesystems.ManagerConfig{
		Default: "local",
		Disks: map[string]filesystems.Disk{
			"local": {
				Driver: "local",
				Config: map[string]interface{}{
					"basePath": "./storage/app",
					"baseURL":  "http://localhost:8080/storage",
				},
			},
			"s3_public": {
				Driver: "s3",
				Config: map[string]interface{}{
					"bucket":    "your-public-s3-bucket",
					"region":    "us-east-1",
					"accessKey": "YOUR_AWS_ACCESS_KEY",
					"secretKey": "YOUR_AWS_SECRET_KEY",
					"baseURL":   "https://your-public-s3-bucket.s3.us-east-1.amazonaws.com",
				},
			},
		},
	}

	// 3. Create the FileSystem manager, providing the logger and trace ID key as options.
	fs, err := filesystems.New(config,
		filesystems.WithLogger(logger),
		filesystems.WithTraceIDKey(TraceIDKey),
	)
	if err != nil {
		logger.Error("Failed to create filesystem manager", "error", err)
		os.Exit(1)
	}

	logger.Info("=== Filesystem Manager Example Starting ===")

	// 4. Create a context with a trace ID.
	traceID := uuid.New().String()
	ctx := context.WithValue(context.Background(), TraceIDKey, traceID)

	logger.InfoContext(ctx, "Starting file operations with trace ID")

	// 5. Use the manager. The trace ID will now be automatically included in all logs.
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
