package database

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// mongoProvider implements the MongoProvider interface.
type mongoProvider struct {
	client *mongo.Client
	db     *mongo.Database
	logger *slog.Logger
}

// NewMongoProvider creates a new MongoDB provider.
// It expects the URI to be fully resolved by the config loader (Viper).
func NewMongoProvider(ctx context.Context, cfg *MongoConfig, policy *PolicyConfig, logger *slog.Logger) (Provider, error) {
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

	if err := provider.Ping(ctx); err != nil {
		return nil, fmt.Errorf("initial MongoDB ping failed: %w", err)
	}

	logger.Info("MongoDB connection successful")
	return provider, nil
}

func (p *mongoProvider) Ping(ctx context.Context) error {
	return p.client.Ping(ctx, readpref.Primary())
}

func (p *mongoProvider) Close() error {
	p.logger.Info("Closing MongoDB connection...")
	return p.client.Disconnect(context.Background())
}

func (p *mongoProvider) Client() interface{} {
	return p.client
}

func (p *mongoProvider) Database() *mongo.Database {
	return p.db
}
