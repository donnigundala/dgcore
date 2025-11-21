package migration

import (
	"context"

	"gorm.io/gorm"
)

// gormStore adalah implementasi Store untuk GORM.
type gormStore struct {
	db *gorm.DB
}

// NewGormStore membuat store baru untuk GORM.
func NewGormStore(db *gorm.DB) Store {
	return &gormStore{db: db}
}

// MigrationModel merepresentasikan skema tabel pelacak migrasi.
type MigrationModel struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"unique;not null"`
}

func (s *gormStore) EnsureMigrationTable(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&MigrationModel{})
}

func (s *gormStore) GetCompletedMigrations(ctx context.Context) ([]string, error) {
	var migrations []MigrationModel
	if err := s.db.WithContext(ctx).Find(&migrations).Error; err != nil {
		return nil, err
	}

	names := make([]string, len(migrations))
	for i, m := range migrations {
		names[i] = m.Name
	}
	return names, nil
}

func (s *gormStore) LogMigration(ctx context.Context, name string) error {
	migration := MigrationModel{Name: name}
	return s.db.WithContext(ctx).Create(&migration).Error
}