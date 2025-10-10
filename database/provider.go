package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	gormprom "gorm.io/plugin/prometheus"

	// Dialectors
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"

	// Observability plugins
	gormotel "gorm.io/plugin/opentelemetry/tracing"
)

// Provider wraps a GORM DB and its underlying *sql.DB with lifecycle helpers.
type Provider struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

// New creates a new database connection using the provided Config.
// It initializes GORM with best-practice defaults, sets pooling,
// pings the DB, and optionally runs automigrations.
func New(cfg Config, modelsForAutoMigrate ...interface{}) (*Provider, error) {
	cfg.Defaults()

	dsn, err := cfg.DSNFor()
	if err != nil {
		return nil, err
	}

	var dialector gorm.Dialector
	switch cfg.Driver {
	case DriverMySQL:
		dialector = mysql.Open(dsn)
	case DriverPostgres:
		dialector = postgres.Open(dsn)
	case DriverSQLite:
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	gormCfg := &gorm.Config{
		SkipDefaultTransaction: cfg.SkipDefaultTransaction,
		PrepareStmt:            cfg.PrepareStmt,
		Logger:                 newGormLogger(cfg.LogLevel, cfg.SlowThreshold),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: cfg.SingularTable,
		},
	}

	var db *gorm.DB
	var sqlDB *sql.DB

	// Retry loop for initial connect + ping
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxConnectRetries; attempt++ {
		db, lastErr = gorm.Open(dialector, gormCfg)
		if lastErr != nil {
			sleep(attempt, cfg.ConnectRetryBackoff)
			continue
		}

		sqlDB, lastErr = db.DB()
		if lastErr != nil {
			sleep(attempt, cfg.ConnectRetryBackoff)
			continue
		}

		// Connection pooling
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

		// Health check ping with timeout
		if lastErr = pingWithTimeout(sqlDB, 5*time.Second); lastErr != nil {
			_ = sqlDB.Close()
			sleep(attempt, cfg.ConnectRetryBackoff)
			continue
		}

		// Connected successfully
		break
	}
	if lastErr != nil {
		return nil, fmt.Errorf("db connect: %w", lastErr)
	}

	// SQLite PRAGMAs for production-friendly behavior
	if cfg.Driver == DriverSQLite {
		if err := applySQLitePragmas(db); err != nil {
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
			return nil, fmt.Errorf("sqlite pragmas: %w", err)
		}
	}

	// Optional observability plugins
	if cfg.EnablePrometheus {
		if db != nil {
			_ = db.Use(gormprom.New(gormprom.Config{
				DBName:          nonEmpty(cfg.DBName, string(cfg.Driver)),
				RefreshInterval: 15, // seconds
				StartServer:     false,
				// Metrics are exposed via prometheus default registry; scrape from app server.
				// You can wire a custom registry if needed.
			}))
		}
	}
	if cfg.EnableOTel {
		if db != nil {
			_ = db.Use(gormotel.NewPlugin(
				gormotel.WithDBSystem(nonEmpty(cfg.DBName, string(cfg.Driver))),
			))
		}
	}

	p := &Provider{Gorm: db, SQL: sqlDB}

	// Optional automigrate (recommended only outside of strict prod flows)
	if cfg.AutoMigrate && len(modelsForAutoMigrate) > 0 {
		if err := p.Migrate(modelsForAutoMigrate...); err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	return p, nil
}

func pingWithTimeout(db *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.PingContext(ctx)
}

// Close gracefully closes the underlying sql.DB.
func (p *Provider) Close() error {
	if p == nil || p.SQL == nil {
		return nil
	}
	return p.SQL.Close()
}

// HealthCheck pings the DB using the provided context (with timeout recommended).
func (p *Provider) HealthCheck(ctx context.Context) error {
	if p == nil || p.SQL == nil {
		return errors.New("db not initialized")
	}
	return p.SQL.PingContext(ctx)
}

// Migrate runs GORM AutoMigrate over provided models.
// Example: provider.Migrate(&User{}, &Order{})
func (p *Provider) Migrate(models ...interface{}) error {
	if p == nil || p.Gorm == nil {
		return errors.New("db not initialized")
	}
	return p.Gorm.AutoMigrate(models...)
}

func sleep(attempt int, base time.Duration) {
	// exponential backoff with jitter could be added; keep simple exponential for now
	delay := base << attempt
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	time.Sleep(delay)
}

func applySQLitePragmas(db *gorm.DB) error {
	// WAL improves concurrency; busy_timeout prevents immediate SQLITE_BUSY errors
	if err := db.Exec("PRAGMA journal_mode = WAL;").Error; err != nil {
		return err
	}
	if err := db.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		return err
	}
	return nil
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
