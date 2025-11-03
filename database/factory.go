package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/donnigundala/dgcore/database/contracts"
	"github.com/donnigundala/dgcore/database/mongo"
	"github.com/donnigundala/dgcore/database/sql"
)

// New creates a new database provider based on the given configuration.
// The appSlog is the application-wide structured logger.
func New(ctx context.Context, cfg *contracts.Config, appSlog *slog.Logger) (contracts.Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is nil")
	}

	metrics := cfg.Metrics
	if metrics == nil {
		metrics = contracts.NewNopMetricsProvider()
	}

	if appSlog == nil {
		appSlog = slog.Default()
	}

	switch cfg.Driver {
	case contracts.ProviderSQL:
		return sql.NewSQLProvider(ctx, cfg.SQL, cfg.Policy, metrics, appSlog, cfg.TraceIDKey)
	case contracts.ProviderMongo:
		return mongo.NewMongoProvider(ctx, cfg.Mongo, cfg.Policy, metrics, appSlog, cfg.TraceIDKey)
	case contracts.ProviderRedis:
		// In the future, a Redis provider can be implemented here.
		return nil, fmt.Errorf("redis provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}
