package migration

import "context"

// Store adalah interface untuk melacak migrasi mana yang telah dijalankan.
// Ini memungkinkan Migrator untuk bekerja dengan berbagai jenis database (SQL, Mongo, dll).
type Store interface {
	// EnsureMigrationTable memastikan bahwa tabel atau collection untuk melacak migrasi sudah ada.
	EnsureMigrationTable(ctx context.Context) error

	// GetCompletedMigrations mengembalikan daftar nama migrasi yang telah selesai.
	GetCompletedMigrations(ctx context.Context) ([]string, error)

	// LogMigration mencatat bahwa sebuah migrasi telah berhasil dijalankan.
	LogMigration(ctx context.Context, name string) error
}