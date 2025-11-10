package database

import (
	"context"
	"fmt"
	"log/slog"
)

// New creates a new database provider based on the given configuration.
// The appSlog is the application-wide structured logger.
func New(ctx context.Context, cfg *Config, appSlog *slog.Logger) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is nil")
	}

	metrics := cfg.Metrics
	if metrics == nil {
		metrics = NewNopMetricsProvider()
	}

	if appSlog == nil {
		appSlog = slog.Default()
	}

	switch cfg.Driver {
	case ProviderSQL:
		return NewSQLProvider(ctx, cfg.SQL, cfg.Policy, metrics, appSlog, cfg.TraceIDKey)
	case ProviderMongo:
		return NewMongoProvider(ctx, cfg.Mongo, cfg.Policy, metrics, appSlog, cfg.TraceIDKey)
	case ProviderRedis:
		// In the future, a Redis provider can be implemented here.
		return nil, fmt.Errorf("redis provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}
