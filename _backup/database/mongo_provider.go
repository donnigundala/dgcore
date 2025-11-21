package database

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/donnigundala/dgcore/ctxutil"
)

// mongoProvider implements the MongoProvider interface.
type mongoProvider struct {
	client *mongo.Client
	db     *mongo.Database
	logger *slog.Logger
}

// newMongoProvider creates a new MongoDB provider.
func newMongoProvider(ctx context.Context, cfg *MongoConfig, policy *PolicyConfig, logger *slog.Logger) (Provider, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("MongoDB URI is missing")
	}

	clientOpts := options.Client().ApplyURI(cfg.URI)
	if cfg.Pool.MaxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(cfg.Pool.MaxPoolSize)
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	provider := &mongoProvider{
		client: client,
		db:     client.Database(cfg.Database),
		logger: logger,
	}

	// Ping the database to ensure the connection is valid.
	if err := provider.Ping(ctx); err != nil {
		return nil, fmt.Errorf("initial MongoDB ping failed: %w", err)
	}

	logger.Info("MongoDB connection successful")
	return provider, nil
}

// Ping verifies the database connection is alive.
func (p *mongoProvider) Ping(ctx context.Context) error {
	log := ctxutil.LoggerFromContext(ctx)
	log.Debug("Pinging MongoDB")
	return p.client.Ping(ctx, readpref.Primary())
}

// Close gracefully terminates the database connection.
func (p *mongoProvider) Close() error {
	p.logger.Info("Closing MongoDB connection...")
	// Use a background context for disconnection as it's a global operation.
	return p.client.Disconnect(context.Background())
}

// Client returns the underlying MongoDB client instance.
func (p *mongoProvider) Client() interface{} {
	return p.client
}

// Database returns the specific mongo.Database instance for this connection.
func (p *mongoProvider) Database() *mongo.Database {
	return p.db
}
