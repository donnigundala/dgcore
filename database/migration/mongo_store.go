package migration

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultMigrationCollection = "migrations"

// mongoStore adalah implementasi Store untuk MongoDB.
type mongoStore struct {
	coll *mongo.Collection
}

// NewMongoStore membuat store baru untuk MongoDB.
func NewMongoStore(db *mongo.Database) Store {
	return &mongoStore{coll: db.Collection(defaultMigrationCollection)}
}

type mongoMigration struct {
	Name string `bson:"name"`
}

func (s *mongoStore) EnsureMigrationTable(ctx context.Context) error {
	// Di MongoDB, collection dibuat secara implisit saat penulisan pertama.
	// Kita bisa membuat index untuk memastikan nama migrasi unik.
	_, err := s.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (s *mongoStore) GetCompletedMigrations(ctx context.Context) ([]string, error) {
	cursor, err := s.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var results []mongoMigration
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	names := make([]string, len(results))
	for i, res := range results {
		names[i] = res.Name
	}
	return names, nil
}

func (s *mongoStore) LogMigration(ctx context.Context, name string) error {
	_, err := s.coll.InsertOne(ctx, mongoMigration{Name: name})
	return err
}