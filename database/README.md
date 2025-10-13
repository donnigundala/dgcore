# GORM DB Provider

A small, production-friendly GORM provider/factory supporting MySQL, PostgreSQL, and SQLite with:
- Sensible connection pooling defaults
- Structured logging with configurable levels and slow-query threshold
- Optional automigrations
- Health checks and graceful shutdown
- Env-friendly configuration

## Install

Add the required dependencies to your project:

- gorm.io/gorm
- gorm.io/driver/postgres
- gorm.io/driver/mysql
- gorm.io/driver/sqlite

## Usage

\`\`\`go
import (
"context"
"time"

"your/module/pkg/db"
)

type User struct {
ID   uint   `gorm:"primaryKey"`
Name string `gorm:"size:120;not null"`
}

func main() {
cfg := db.Config{
Driver:       db.DriverPostgres,
Host:         "localhost",
Port:         5432,
User:         "postgres",
Password:     "password",
DBName:       "app",
SSLMode:      "disable",
TimeZone:     "UTC",
AutoMigrate:  true,
LogLevel:     db.LogWarn,
SlowThreshold: 250 * time.Millisecond,
SkipDefaultTransaction: true,
}
cfg.Defaults()

provider, err := db.New(cfg, &User{})
if err != nil {
panic(err)
}
defer provider.Close()

// Health check
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
if err := provider.HealthCheck(ctx); err != nil {
panic(err)
}

// Use provider.Gorm for queries
_ = provider.Gorm.Create(&User{Name: "Alice"}).Error
}
\`\`\`

## Environment Variables

You can also load configuration via env:

\`\`\`
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=app
DB_SSLMODE=disable
DB_TIMEZONE=UTC
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m
DB_SKIP_DEFAULT_TX=true
DB_PREPARE_STMT=false
DB_SINGULAR_TABLE=false
DB_LOG_LEVEL=warn
DB_SLOW_THRESHOLD=200ms
DB_AUTO_MIGRATE=false
DB_MAX_CONNECT_RETRIES=5
DB_CONNECT_RETRY_BACKOFF=500ms
DB_ENABLE_PROMETHEUS=false
DB_PROM_NAMESPACE=gorm
DB_ENABLE_OTEL=false
\`\`\`

\`\`\`go
cfg, _ := db.LoadFromEnv("DB")
provider, err := db.New(cfg /*, models... */)
if err != nil {
panic(err)
}
defer provider.Close()
\`\`\`

## Observability

- Prometheus metrics: enable via `DB_ENABLE_PROMETHEUS=true`. The provider registers the GORM Prometheus plugin with a refresh interval and uses the default registry. Scrape metrics from your app's metrics endpoint.
- OpenTelemetry tracing: enable via `DB_ENABLE_OTEL=true`. The provider registers the official `gorm.io/plugin/opentelemetry/tracing` plugin with default options.

Note on versions:
- Some versions of the OTEL plugin support `tracing.WithDBName(...)`. If your build includes it, you may optionally pass it:
  \`\`\`go
  _ = provider.Gorm.Use(tracing.NewPlugin(tracing.WithDBName("appdb")))
  \`\`\`
  If your version does not expose `WithDBName`, use `tracing.NewPlugin()` without options (as implemented by default).

## SQLite Notes

SQLite is initialized with production-friendly PRAGMAs:
- WAL mode for better concurrency
- `PRAGMA busy_timeout = 5000` to reduce SQLITE_BUSY errors

You can set `DB_SQLITE_PATH` (e.g., `file:app.db`, `:memory:`) and append custom params via `Config.Params`.

## MySQL and Postgres Notes

- MySQL DSN includes sane defaults: `charset=utf8mb4`, `parseTime=True`, `loc=UTC`, and timeouts.
- Postgres DSN includes `sslmode` and `TimeZone` with defaults (`disable`, `UTC`). For production, consider `require`/`verify-full` and appropriate CA configuration.
