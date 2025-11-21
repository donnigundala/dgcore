package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/database"
	"github.com/spf13/cobra"
)

type WipeCommand struct {
	force      bool
	connection string
}

func (c *WipeCommand) Signature() string {
	return "db:wipe"
}

func (c *WipeCommand) Description() string {
	return "Drop all tables/collections from the database"
}

func (c *WipeCommand) Configure(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&c.force, "force", "f", false, "Force the wipe operation")
	cmd.Flags().StringVar(&c.connection, "connection", "", "The database connection to wipe")
}

func (c *WipeCommand) Handle(cmd *cobra.Command, args []string) error {
	if !c.force {
		fmt.Println("This is a destructive command. Use --force to proceed.")
		return nil
	}

	// --- Bootstrap Aplikasi ---
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	config.Load()

	var dbManagerConfig database.Config
	if err := config.Inject("databases", &dbManagerConfig); err != nil {
		logger.Error("Failed to inject database configurations", "error", err)
		return err
	}

	// Jika nama koneksi tidak diberikan via flag, gunakan default dari config
	if c.connection == "" {
		c.connection = dbManagerConfig.DefaultConnection
		if c.connection == "" {
			logger.Error("No connection specified and no default connection found in config.")
			return fmt.Errorf("no connection specified")
		}
	}

	dbManager, err := database.NewManager(dbManagerConfig, database.WithLogger(logger))
	if err != nil {
		logger.Error("Failed to create database manager", "error", err)
		return err
	}
	defer dbManager.Close()

	logger.Info("Attempting to wipe database...", "connection", c.connection)

	provider, err := dbManager.Connection(c.connection)
	if err != nil {
		logger.Error("Failed to get database connection", "connection", c.connection, "error", err)
		return err
	}

	// --- Lakukan Wipe Berdasarkan Tipe Driver ---
	switch p := provider.(type) {
	case database.SQLProvider:
		gormDB := p.GormWithContext(context.Background())
		logger.Info("Wiping SQL database...")

		// Dapatkan semua nama tabel
		var tableNames []string
		if err := gormDB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = ?", "public").Scan(&tableNames).Error; err != nil {
			logger.Error("Failed to list tables for wiping", "error", err)
			return err
		}

		// Hapus semua tabel (kecuali tabel migrasi)
		for _, table := range tableNames {
			if table == "migration_models" { // Sesuaikan dengan nama tabel migrasi GORM Anda
				continue
			}
			logger.Info("Dropping table", "table", table)
			if err := gormDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
				logger.Error("Failed to drop table", "table", table, "error", err)
			}
		}
		logger.Info("✅ SQL database wipe complete.")

	case database.MongoProvider:
		mongoDB := p.Database()
		logger.Info("Wiping MongoDB database...", "database", mongoDB.Name())
		if err := mongoDB.Drop(context.Background()); err != nil {
			logger.Error("Failed to drop MongoDB database", "error", err)
			return err
		}
		logger.Info("✅ MongoDB database wipe complete.")
	default:
		logger.Error("Unsupported database provider for wipe operation", "connection", c.connection)
		return fmt.Errorf("unsupported provider")
	}

	return nil
}
