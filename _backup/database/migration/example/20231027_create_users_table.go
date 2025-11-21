package example

import (
	"context"

	"gorm.io/gorm"
)

// --- Definisikan Model Anda di sini atau import dari package lain ---

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"unique"`
}

// --- Definisikan Fungsi Migrasi Anda ---
// Setiap fungsi migrasi harus sesuai dengan signature `migration.MigrationFunc`.

// CreateUsersTable membuat tabel users.
func CreateUsersTable(ctx context.Context, db any) error {
	// Lakukan type assertion ke *gorm.DB
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("CreateUsersTable expects a *gorm.DB connection")
	}
	return gormDB.WithContext(ctx).AutoMigrate(&User{})
}

// AddEmailToUsersTable menambahkan kolom email ke tabel users.
func AddEmailToUsersTable(ctx context.Context, db any) error {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("AddEmailToUsersTable expects a *gorm.DB connection")
	}
	return gormDB.WithContext(ctx).Migrator().AddColumn(&User{}, "Email")
}
