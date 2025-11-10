# `core/cache` Package

## Overview

The `cache` package provides a unified interface for interacting with various caching backends like Redis and Memcache. It is designed for high performance and ease of use, integrating seamlessly with the framework's configuration and logging systems.

## Features

- **Unified Interface**: A single `Provider` interface for all cache operations, allowing you to switch backends with only a configuration change.
- **Manager Pattern**: A singleton `CacheManager` handles the lifecycle of multiple named cache connections, consistent with the `database` package.
- **Framework-Integrated Configuration**: Leverages the central `config` package, allowing configuration via YAML files and environment variables with sensible defaults.
- **Structured Logging**: Integrates with `slog` for consistent, structured logging.
- **Built-in Providers**: Comes with ready-to-use providers for Redis and Memcache.

## Configuration

The cache manager is configured through a `caches` block in your application's `config.yaml`. You can define multiple named connections.

The configuration loading follows this order of precedence (highest to lowest):
1.  **Environment Variables**
2.  **`config.yaml` File**
3.  **Framework Defaults**

### YAML Configuration

```yaml
# your-app/config/config.yaml
caches:
  # The 'default' cache connection
  default:
    driver: "redis"
    namespace: "my_app"
    separator: ":"
    redis:
      host: "127.0.0.1"
      port: "6379"
      password: ${REDIS_PASSWORD} # Example of using env var inside yaml
      db: 0
      ttl: "2h"

  # A second cache connection using Memcache
  session_cache:
    driver: "memcache"
    namespace: "sessions"
    memcache:
      servers:
        - "127.0.0.1:11211"
        - "127.0.0.1:11212"
      ttl: "30m"
```

### Environment Variable Configuration

You can override any setting using environment variables. The key is constructed from the YAML path.

```sh
# Override the default Redis host and port
export CACHES_DEFAULT_REDIS_HOST="redis.example.com"
export CACHES_DEFAULT_REDIS_PORT="6380"

# Override the Memcache servers for the session_cache
export CACHES_SESSION_CACHE_MEMCACHE_SERVERS="mem1.example.com:11211,mem2.example.com:11211"
```

## Usage

The idiomatic workflow involves using the `CacheManager` to get a specific cache provider.

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/donnigundala/dgcore/cache"
    "github.com/donnigundala/dgcore/config"
)

func main() {
    // 1. Bootstrap logger and config
    appSlog := slog.New(slog.NewTextHandler(os.Stdout, nil))
    config.Load()

    // 2. Initialize the Cache Manager
    cacheManager := cache.Manager()
    cacheManager.SetLogger(appSlog)

    var cacheConfigs map[string]*cache.Config
    config.Inject("caches", &cacheConfigs)

    for name, cfg := range cacheConfigs {
        cacheManager.Register(name, cfg)
    }
    cacheManager.ConnectAll(context.Background())
    defer cacheManager.Close()

    // 3. Get a specific cache provider
    redisCache := cacheManager.Get("default")
    if redisCache == nil {
        // handle error
    }

    // 4. Use the cache
    ctx := context.Background()
    redisCache.Set(ctx, "my_key", "my_value")

    var result string
    redisCache.Get(ctx, "my_key", &result)
    slog.Info("Got value from cache", "value", result)
}
```