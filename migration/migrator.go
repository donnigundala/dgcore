package migration

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"
)

// MigrationFunc adalah tipe fungsi untuk sebuah migrasi.
// Ia menerima konteks dan koneksi database.
type MigrationFunc func(ctx context.Context, db any) error

// Migration adalah struct yang membungkus fungsi migrasi dengan nama unik.
type Migration struct {
	Name string
	Up   MigrationFunc
}

// Migrator mengelola dan menjalankan migrasi database.
type Migrator struct {
	db        any // Koneksi database (bisa *gorm.DB, *mongo.Database, dll.)
	logger    *slog.Logger
	store     Store
	migrations map[string]MigrationFunc
}

// NewMigrator membuat instance Migrator baru.
// Ia menerima koneksi database, logger, dan store untuk melacak status migrasi.
func NewMigrator(db any, logger *slog.Logger, store Store) *Migrator {
	return &Migrator{
		db:        db,
		logger:    logger.With("component", "migrator"),
		store:     store,
		migrations: make(map[string]MigrationFunc),
	}
}

// Register menambahkan sebuah migrasi ke dalam daftar untuk dijalankan.
// Nama migrasi harus unik.
func (m *Migrator) Register(name string, up MigrationFunc) {
	if _, exists := m.migrations[name]; exists {
		m.logger.Warn("Migration with this name already registered, overwriting.", "name", name)
	}
	m.migrations[name] = up
}

// RunAll menjalankan semua migrasi yang terdaftar yang belum dijalankan.
// Migrasi dijalankan secara berurutan berdasarkan nama (alphanumerical).
func (m *Migrator) RunAll(ctx context.Context) error {
	m.logger.Info("Starting database migration process...")

	if err := m.store.EnsureMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migration tracking table/collection exists: %w", err)
	}

	completedMigrations, err := m.store.GetCompletedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get completed migrations: %w", err)
	}

	completedMap := make(map[string]bool)
	for _, mig := range completedMigrations {
		completedMap[mig] = true
	}

	// Urutkan nama migrasi agar eksekusi deterministik
	var pendingMigrations []string
	for name := range m.migrations {
		if !completedMap[name] {
			pendingMigrations = append(pendingMigrations, name)
		}
	}
	sort.Strings(pendingMigrations)

	if len(pendingMigrations) == 0 {
		m.logger.Info("No new migrations to run. Database is up to date.")
		return nil
	}

	m.logger.Info("Found pending migrations to run.", "count", len(pendingMigrations), "migrations", pendingMigrations)

	for _, name := range pendingMigrations {
		m.logger.Info("Running migration...", "name", name)
		
		start := time.Now()
		migrationFunc := m.migrations[name]

		// Di masa depan, ini bisa dibungkus dalam transaksi jika store mendukungnya.
		if err := migrationFunc(ctx, m.db); err != nil {
			m.logger.Error("Migration failed", "name", name, "error", err, "duration", time.Since(start))
			return fmt.Errorf("migration '%s' failed: %w", name, err)
		}

		if err := m.store.LogMigration(ctx, name); err != nil {
			m.logger.Error("Failed to log completed migration", "name", name, "error", err)
			// Ini adalah error kritis karena bisa menyebabkan migrasi dijalankan lagi.
			return fmt.Errorf("failed to log completion for migration '%s': %w", name, err)
		}

		m.logger.Info("Migration completed successfully", "name", name, "duration", time.Since(start))
	}

	m.logger.Info("✅ All migrations completed successfully.")
	return nil
}