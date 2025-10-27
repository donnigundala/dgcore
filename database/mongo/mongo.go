package mongo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Connector struct {
	client *mongo.Client
	config *Config
}

func New(cfg *Config) (*Connector, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := cfg.URI
	if uri == "" {
		// Build hosts (multi-host support)
		hosts := strings.Join(cfg.Hosts, ",")
		scheme := "mongodb"
		if cfg.UseSRV {
			scheme = "mongodb+srv"
		}

		uri = fmt.Sprintf("%s://%s:%s@%s/%s?authSource=%s",
			scheme,
			cfg.Username,
			cfg.Password,
			hosts,
			cfg.Database,
			cfg.AuthSource,
		)

		if cfg.ReplicaSet != "" {
			uri += fmt.Sprintf("&replicaSet=%s", cfg.ReplicaSet)
		}
	}

	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(uint64(cfg.MaxPoolSize)).
		SetMinPoolSize(uint64(cfg.MinPoolSize)).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetServerSelectionTimeout(cfg.ServerSelectionTTL)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("MONGODB: connect failed: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MONGODB: ping failed: %w", err)
	}

	log.Printf("[DATABASE] Connected to MongoDB URIs(%s)", strings.Join(cfg.Hosts, ","))
	return &Connector{client: client, config: cfg}, nil
}

// Connect returns the underlying mongo.Client
func (c *Connector) Client() *mongo.Client {
	return c.client
}

// Close database connection
func (c *Connector) Close() error {
	if c.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// Ping database to check connection health
func (c *Connector) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}
