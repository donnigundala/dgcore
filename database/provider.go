package database

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Provider defines the common interface for all database connections.
// This allows for treating different database systems (SQL, MongoDB, etc.)
// in a uniform manner.
type Provider interface {
	// Ping checks if the database connection is alive.
	Ping(ctx context.Context) error

	// Close gracefully terminates the database connection.
	Close() error
}

// SQLProvider extends the base Provider with methods specific to SQL databases,
// typically those provided by a library like GORM.
type SQLProvider interface {
	Provider
	// Gorm returns the underlying GORM DB instance for complex queries.
	// DEPRECATED: This method is not context-aware. Use GormWithContext instead.
	Gorm() interface{}
	// GormWithContext returns a new GORM session bound to the provided context.
	GormWithContext(ctx context.Context) *gorm.DB
}

// MongoProvider extends the base Provider with methods specific to MongoDB.
type MongoProvider interface {
	Provider
	// Client returns the underlying MongoDB client instance.
	// It returns interface{} to avoid a direct Mongo driver dependency in this package.
	Client() interface{}
	// Database returns the specific mongo.Database instance for this connection.
	Database() *mongo.Database
}

// newProvider creates a new database provider based on the given configuration.
// This factory is the central point for dispatching to the correct provider constructor.
func newProvider(ctx context.Context, name string, cfg Connection, appSlog *slog.Logger) (Provider, error) {
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
		return newSQLProvider(ctx, cfg.SQL, &cfg.Policy, connLogger)

	case "mongo":
		if cfg.Mongo == nil {
			return nil, fmt.Errorf("mongo config is missing for driver 'mongo' on connection '%s'", name)
		}
		return newMongoProvider(ctx, cfg.Mongo, &cfg.Policy, connLogger)

	default:
		return nil, fmt.Errorf("unsupported driver '%s' for connection '%s'", cfg.Driver, name)
	}
}
