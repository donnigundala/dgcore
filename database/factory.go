package database

import (
	"context"
	"fmt"
	"log/slog"
)

// New creates a new database provider based on the given configuration.
// This factory is the central point for dispatching to the correct provider constructor.
func New(ctx context.Context, name string, cfg Connection, appSlog *slog.Logger) (Provider, error) {
	if appSlog == nil {
		// Fallback to a default logger if none is provided.
		appSlog = slog.Default()
	}

	// Create a logger specific to this connection for better traceability.
	connLogger := appSlog.With("connection", name, "driver", cfg.Driver)

	switch cfg.Driver {
	case "sql":
		if cfg.SQL == nil {
			return nil, fmt.Errorf("SQL config is missing for driver 'sql' on connection '%s'", name)
		}
		return NewSQLProvider(ctx, cfg.SQL, &cfg.Policy, connLogger)

	case "mongo":
		if cfg.Mongo == nil {
			return nil, fmt.Errorf("mongo config is missing for driver 'mongo' on connection '%s'", name)
		}
		return NewMongoProvider(ctx, cfg.Mongo, &cfg.Policy, connLogger)

	default:
		return nil, fmt.Errorf("unsupported driver '%s' for connection '%s'", cfg.Driver, name)
	}
}
