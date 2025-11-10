package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/database"
	"github.com/donnigundala/dgcore/migration"
	"github.com/donnigundala/dgcore/migration/example" // Import package yang berisi fungsi migrasi
	"gorm.io/gorm"
)

func main() {
	// 1. Bootstrap Aplikasi (Logger, Config, DB Manager)
	appSlog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(appSlog)

	// Muat konfigurasi dari file (misal: config/database.yaml) dan env vars
	config.Load()

	// Unmarshal bagian 'databases' dari config ke struct ManagerConfig.
	var dbManagerConfig database.ManagerConfig
	if err := config.Inject("databases", &dbManagerConfig); err != nil {
		appSlog.Error("Failed to inject database configurations", "error", err)
		os.Exit(1)
	}

	// Buat DatabaseManager dengan Dependency Injection
	dbManager, err := database.NewManager(dbManagerConfig, database.WithLogger(appSlog))
	if err != nil {
		appSlog.Error("Failed to create database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// 2. Dapatkan Koneksi DB dan Buat Migrator
	// Ganti "my_postgres" dengan nama koneksi yang ingin Anda migrasikan dari file config.
	connName := "my_postgres"
	provider, err := dbManager.Connection(connName)
	if err != nil {
		appSlog.Error("Failed to get database connection", "connection", connName, "error", err)
		os.Exit(1)
	}

	// Lakukan type assertion untuk mendapatkan instance DB spesifik (*gorm.DB)
	sqlProvider, ok := provider.(database.SQLProvider)
	if !ok {
		appSlog.Error("Connection is not a SQL provider", "connection", connName)
		os.Exit(1)
	}
	gormDB := sqlProvider.Gorm().(*gorm.DB)

	// Buat store dan migrator yang sesuai untuk GORM
	store := migration.NewGormStore(gormDB)
	migrator := migration.NewMigrator(gormDB, appSlog, store)

	// 3. Register (Daftarkan) Migrasi Anda secara Eksplisit
	// Gunakan timestamp atau nomor urut pada nama untuk memastikan urutan eksekusi yang benar.
	// Nama ini yang akan disimpan di tabel 'migrations'.
	migrator.Register("20231027100000_create_users_table", example.CreateUsersTable)
	migrator.Register("20231027100500_add_email_to_users_table", example.AddEmailToUsersTable)
	// ... daftarkan migrasi lainnya di sini

	// 4. Jalankan Semua Migrasi yang Tertunda
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := migrator.RunAll(ctx); err != nil {
		appSlog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	// Jika berhasil, pesan ini akan muncul.
	// appSlog.Info("✅ All migrations completed successfully.") // Pesan ini sudah ada di dalam RunAll()
}
