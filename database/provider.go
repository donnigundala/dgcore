package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
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
// typically those provided by GORM.
type SQLProvider interface {
	Provider
	// Gorm returns the underlying GORM DB instance for complex queries.
	Gorm() interface{} // Returning interface{} to avoid direct GORM dependency here.
}

// MongoProvider extends the base Provider with methods specific to MongoDB.
type MongoProvider interface {
	Provider
	// Client returns the underlying MongoDB client instance.
	Client() interface{} // Returning interface{} to avoid direct Mongo driver dependency here.
	Database() *mongo.Database
}
