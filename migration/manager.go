package migration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/donnigundala/dgcore/database"
	mongomigration "github.com/donnigundala/dgcore/migration/mongo"
	sqlmigration "github.com/donnigundala/dgcore/migration/sql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

var ( 
    // registeredMigrations holds all migrations registered via init() functions.
    registeredMigrations []Migration
)

// Register should be called by migration files in their init() function.
func Register(m Migration) {
	registeredMigrations = append(registeredMigrations, m)
}

// Migration defines a common interface for migrations.
type Migration interface {
	Name() string
}

// MongoMigration defines a mongo migration.
type MongoMigration interface {
	Migration
	Up() error
	Down() error
}

// SQLMigration defines a sql migration.
type SQLMigration interface {
	Migration
	Up(tx *gorm.DB) error
	Down(tx *gorm.DB) error
}

// Manager manages migrations.
type Manager struct {
	mongo *mongo.Client
	gorm  *gorm.DB
	logger *slog.Logger

	MongoMigrations []MongoMigration // Made public
	SQLMigrations   []SQLMigration   // Made public
}

// NewManagerFromDB creates a new manager from a database manager.
func NewManagerFromDB(dbManager *database.DatabaseManager, logger *slog.Logger) (*Manager, error) {
	sqlProvider := dbManager.Get("default_sql")
	if sqlProvider == nil {
		return nil, errors.New("sql provider not found")
	}

	gormDB, ok := sqlProvider.GetWriter().(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("invalid sql provider type: %T", sqlProvider.GetWriter())
	}

	mongoProvider := dbManager.Get("default_mongo")
	if mongoProvider == nil {
		return nil, errors.New("mongo provider not found")
	}

	mongoClient, ok := mongoProvider.GetWriter().(*mongo.Client)
	if !ok {
		return nil, fmt.Errorf("invalid mongo provider type: %T", mongoProvider.GetWriter())
	}

	m := NewManager(mongoClient, gormDB, logger)
	return m, nil
}

// NewManager creates a new manager.
func NewManager(
	mongo *mongo.Client,
	gorm *gorm.DB,
	logger *slog.Logger,
) *Manager {
	m := &Manager{
		mongo: mongo,
		gorm:  gorm,
		logger: logger,
	}

	// Populate migrations from the global registry.
	for _, mig := range registeredMigrations {
		if sqlMig, ok := mig.(SQLMigration); ok {
			m.SQLMigrations = append(m.SQLMigrations, sqlMig)
		}
		if mongoMig, ok := mig.(MongoMigration); ok {
			m.MongoMigrations = append(m.MongoMigrations, mongoMig)
		}
	}

	return m
}


// Up migrates up.
func (m *Manager) Up() error {
	m.logger.Info("Starting migrations...")

	if err := m.gorm.AutoMigrate(&sqlmigration.Migration{}); err != nil {
		return err
	}

	collection := m.mongo.Database("dg-framework").Collection("migrations")

	var sqlBatch int
	if err := m.gorm.Model(&sqlmigration.Migration{}).Select("max(batch)").Row().Scan(&sqlBatch); err != nil {
		// Handle case where table is empty or other errors
		sqlBatch = 0
	}

	var mongoBatch int
	var result mongomigration.Migration
	opts := options.FindOne().SetSort(bson.D{{"batch", -1}})
	if err := collection.FindOne(context.Background(), bson.D{}, opts).Decode(&result); err == nil {
		mongoBatch = result.Batch
	}

	batch := max(sqlBatch, mongoBatch) + 1
	m.logger.Info("Running migrations for batch", "batch", batch)

	for _, migration := range m.MongoMigrations {
		var result mongomigration.Migration
		if err := collection.FindOne(context.Background(), bson.M{"name": migration.Name()}).Decode(&result); err == nil {
			m.logger.Info("Skipping already applied mongo migration", "name", migration.Name())
			continue
		}

		m.logger.Info("Applying mongo migration", "name", migration.Name())
		if err := migration.Up(); err != nil {
			return err
		}

		if _, err := collection.InsertOne(context.Background(), &mongomigration.Migration{Name: migration.Name(), Batch: batch}); err != nil {
			return err
		}
		m.logger.Info("Applied mongo migration", "name", migration.Name())
	}

	for _, migration := range m.SQLMigrations {
		tx := m.gorm.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		var count int64
		tx.Model(&sqlmigration.Migration{}).Where("name = ?", migration.Name()).Count(&count)
		if count > 0 {
			m.logger.Info("Skipping already applied sql migration", "name", migration.Name())
			tx.Rollback()
			continue
		}

		m.logger.Info("Applying sql migration", "name", migration.Name())
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Create(&sqlmigration.Migration{Name: migration.Name(), Batch: batch}).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			return err
		}
		m.logger.Info("Applied sql migration", "name", migration.Name())
	}

	m.logger.Info("Migrations finished.")
	return nil
}

// Down migrates down.
func (m *Manager) Down() error {
	m.logger.Info("Starting rollbacks...")

	if err := m.gorm.AutoMigrate(&sqlmigration.Migration{}); err != nil {
		return err
	}

	collection := m.mongo.Database("dg-framework").Collection("migrations")

	var sqlBatch int
	if err := m.gorm.Model(&sqlmigration.Migration{}).Select("max(batch)").Row().Scan(&sqlBatch); err != nil {
		// Handle case where table is empty or other errors
		sqlBatch = 0
	}

	var mongoBatch int
	var result mongomigration.Migration
	opts := options.FindOne().SetSort(bson.D{{"batch", -1}})
	if err := collection.FindOne(context.Background(), bson.D{}, opts).Decode(&result); err == nil {
		mongoBatch = result.Batch
	}

	batch := max(sqlBatch, mongoBatch)
	if batch == 0 {
		m.logger.Info("No migrations to roll back.")
		return nil
	}
	m.logger.Info("Rolling back migrations for batch", "batch", batch)

	for i := len(m.MongoMigrations) - 1; i >= 0; i-- {
		migration := m.MongoMigrations[i]

		var result mongomigration.Migration
		if err := collection.FindOne(context.Background(), bson.M{"name": migration.Name(), "batch": batch}).Decode(&result); err != nil {
			continue
		}

		m.logger.Info("Rolling back mongo migration", "name", migration.Name())
		if err := migration.Down(); err != nil {
			return err
		}

		if _, err := collection.DeleteOne(context.Background(), bson.M{"name": migration.Name()}); err != nil {
			return err
		}
		m.logger.Info("Rolled back mongo migration", "name", migration.Name())
	}

	for i := len(m.SQLMigrations) - 1; i >= 0; i-- {
		migration := m.SQLMigrations[i]

		tx := m.gorm.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		var mRec sqlmigration.Migration
		if err := tx.Model(&sqlmigration.Migration{}).Where("name = ? AND batch = ?", migration.Name(), batch).First(&mRec).Error; err != nil {
			tx.Rollback()
			continue
		}

		m.logger.Info("Rolling back sql migration", "name", migration.Name())
		if err := migration.Down(tx); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Where("name = ?", migration.Name()).Delete(&sqlmigration.Migration{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			return err
		}
		m.logger.Info("Rolled back sql migration", "name", migration.Name())
	}

	m.logger.Info("Rollbacks finished.")
	return nil
}
