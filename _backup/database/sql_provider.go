package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"github.com/donnigundala/dgcore/ctxutil"
)

// sqlProvider implements the SQLProvider interface using GORM.
type sqlProvider struct {
	db     *gorm.DB
	logger *slog.Logger
}

// newSQLProvider creates a new GORM-based SQL provider.
func newSQLProvider(ctx context.Context, cfg *SQLConfig, policy *PolicyConfig, logger *slog.Logger) (Provider, error) {
	dsn := buildDSN(&cfg.Primary, cfg.DriverName)
	if dsn == "" {
		return nil, fmt.Errorf("could not build DSN for driver: %s", cfg.DriverName)
	}

	var dialector gorm.Dialector
	switch cfg.DriverName {
	case "postgres", "pgx":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported SQL driver: %s", cfg.DriverName)
	}

	// Configure GORM logger to use our slog-based logger.
	gormLogLevel := gormlogger.Warn // Default level
	if level, ok := gormLogMapping[cfg.LogLevel]; ok {
		gormLogLevel = level
	}

	newLogger := gormlogger.New(
		slog.NewLogLogger(logger.Handler(), slog.Level(gormLogLevel)),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// --- Configure Read/Write Splitting ---
	if len(cfg.Replicas) > 0 {
		logger.Info("Configuring read replicas...", "replica_count", len(cfg.Replicas))
		replicaDialectors := make([]gorm.Dialector, len(cfg.Replicas))
		for i, replicaCfg := range cfg.Replicas {
			replicaDSN := buildDSN(&replicaCfg, cfg.DriverName)
			switch cfg.DriverName {
			case "postgres", "pgx":
				replicaDialectors[i] = postgres.Open(replicaDSN)
			case "mysql":
				replicaDialectors[i] = mysql.Open(replicaDSN)
			case "sqlite":
				replicaDialectors[i] = sqlite.Open(replicaDSN)
			}
		}

		resolver := dbresolver.Register(dbresolver.Config{
			Replicas: replicaDialectors,
			Policy:   dbresolver.RandomPolicy{}, // Use random load balancing for replicas
		})

		if err := db.Use(resolver); err != nil {
			return nil, fmt.Errorf("failed to configure dbresolver: %w", err)
		}
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Pool.ConnMaxLifetime)

	provider := &sqlProvider{
		db:     db,
		logger: logger,
	}

	// Ping the database to ensure the connection is valid.
	if err := provider.Ping(ctx); err != nil {
		return nil, fmt.Errorf("initial database ping failed: %w", err)
	}

	logger.Info("SQL database connection successful")
	return provider, nil
}

// Ping verifies the database connection is alive.
func (p *sqlProvider) Ping(ctx context.Context) error {
	log := ctxutil.LoggerFromContext(ctx)
	log.Debug("Pinging SQL database")

	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close gracefully terminates the database connection.
func (p *sqlProvider) Close() error {
	p.logger.Info("Closing SQL database connection...")
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Gorm returns the underlying GORM DB instance.
// IMPORTANT: This method is now DEPRECATED in favor of GormWithContext.
// It returns a session that is not context-aware.
func (p *sqlProvider) Gorm() interface{} {
	return p.db
}

// GormWithContext returns a new GORM session that is bound to the provided context.
// This is the recommended way to get a GORM instance for database operations,
// as it ensures that logging, tracing, and cancellation are correctly propagated.
func (p *sqlProvider) GormWithContext(ctx context.Context) *gorm.DB {
	// Use WithContext to create a new session that carries the context.
	// GORM's logger will use this context to extract request-specific values.
	return p.db.WithContext(ctx)
}

// buildDSN constructs the Data Source Name from the config.
func buildDSN(node *NodeConfig, driverName string) string {
	p := node
	switch driverName {
	case "postgres", "pgx":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			p.Host, p.Port, p.User, p.Password, p.DBName)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			p.User, p.Password, p.Host, p.Port, p.DBName)
	case "sqlite":
		return p.DBName
	default:
		return ""
	}
}

// gormLogMapping maps our log level strings to GORM's log levels.
var gormLogMapping = map[string]gormlogger.LogLevel{
	"silent": gormlogger.Silent,
	"error":  gormlogger.Error,
	"warn":   gormlogger.Warn,
	"info":   gormlogger.Info,
}
