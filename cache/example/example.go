package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/donnigundala/dgcore/cache"
	"github.com/donnigundala/dgcore/config"
)

// This example assumes you have a `config` directory next to your executable
// with a `cache.yaml` file inside it.

func main() {
	// 1. Initialize a logger for the application.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 2. Load all configurations from the default path ("./config").
	// This will pick up `cache.yaml`, `app.yaml`, etc., and merge them.
	// It also handles .env files and environment variable overrides.
	config.Load()

	// 3. Define a struct to hold the configurations for all caches.
	// The top-level key in your YAML file should be "caches".
	var cacheConfigs map[string]*cache.Config
	if err := config.Inject("caches", &cacheConfigs); err != nil {
		slog.Error("Failed to inject cache configuration", "error", err)
		os.Exit(1)
	}

	// 4. Create a new Manager by injecting the loaded configurations.
	// This is the new Dependency Injection pattern.
	manager, err := cache.New(cacheConfigs, cache.WithLogger(logger))
	if err != nil {
		slog.Error("Failed to create cache manager", "error", err)
		os.Exit(1)
	}
	defer manager.Close()

	slog.Info("Cache manager initialized successfully.")

	// 5. Get a specific cache provider.
	// Using MustGet for convenience; it panics if the disk is not found.
	// Use Get() for safe error handling.
	redisCache := manager.MustGet("default")
	memcacheCache := manager.MustGet("session")

	// 6. Use the cache providers.
	ctx := context.Background()

	// --- Using the 'default' (Redis) cache ---
	slog.Info("--- Using Redis Cache ---")
	myKey := "user:123"
	myValue := map[string]any{"name": "Donni", "level": 99}

	slog.Info("Setting value in Redis...", "key", myKey)
	if err := redisCache.Set(ctx, myKey, myValue, 10*time.Second); err != nil {
		slog.Error("Failed to set value in Redis", "error", err)
	}

	var retrievedValue map[string]any
	slog.Info("Getting value from Redis...", "key", myKey)
	if err := redisCache.Get(ctx, myKey, &retrievedValue); err != nil {
		slog.Error("Failed to get value from Redis", "error", err)
	}
	fmt.Printf("Retrieved from Redis: %+v\n", retrievedValue)

	// --- Using the 'session' (Memcache) cache ---
	slog.Info("\n--- Using Memcache Cache ---")
	sessionKey := "sess:abcde12345"
	sessionValue := "user_id:123"

	slog.Info("Setting value in Memcache...", "key", sessionKey)
	if err := memcacheCache.Set(ctx, sessionKey, sessionValue, 5*time.Second); err != nil {
		slog.Error("Failed to set value in Memcache", "error", err)
	}

	var retrievedSession string
	slog.Info("Getting value from Memcache...", "key", sessionKey)
	if err := memcacheCache.Get(ctx, sessionKey, &retrievedSession); err != nil {
		slog.Error("Failed to get value from Memcache", "error", err)
	}
	fmt.Printf("Retrieved from Memcache: %s\n", retrievedSession)

	// Wait a moment to show TTL works
	slog.Info("\nWaiting for 6 seconds to test Memcache TTL...")
	time.Sleep(6 * time.Second)

	var expiredSession string
	if err := memcacheCache.Get(ctx, sessionKey, &expiredSession); err != nil {
		slog.Error("Error checking expired key", "error", err)
	}

	if expiredSession == "" {
		slog.Info("Value in Memcache has correctly expired.", "key", sessionKey)
	} else {
		slog.Warn("Value in Memcache should have expired but was found.", "value", expiredSession)
	}
}
