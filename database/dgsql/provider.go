package dgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	gormotel "gorm.io/plugin/opentelemetry/tracing"
	gormprom "gorm.io/plugin/prometheus"
)

type Provider struct {
	db  *gorm.DB
	sql *sql.DB
}

type IProvider interface {
	DB() *gorm.DB
	SQL() *sql.DB
	Close() error
	HealthCheck(ctx context.Context) error
	Migrate(models ...interface{}) error
}

func New(cfg *Config, modelsForAutoMigrate ...interface{}) (*Provider, error) {
	dsn, err := cfg.DSNFor()
	if err != nil {
		return nil, err
	}

	var dialector gorm.Dialector
	switch cfg.Driver {
	case DriverMySQL:
		dialector = mysql.Open(dsn)
	case DriverPostgres, DriverPgSQL:
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

	p := &Provider{db: db, sql: sqlDB}

	// Optional automigrate (recommended only outside of strict prod flows)
	if cfg.AutoMigrate && len(modelsForAutoMigrate) > 0 {
		if err := p.Migrate(modelsForAutoMigrate...); err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	return p, nil
}

// DB returns the underlying *gorm.DB instance.
func (p *Provider) DB() *gorm.DB {
	return p.db
}

// SQL returns the underlying *sql.DB instance.
func (p *Provider) SQL() *sql.DB {
	return p.sql
}

// Close gracefully closes the underlying sql.DB.
func (p *Provider) Close() error {
	if p == nil || p.sql == nil {
		return nil
	}
	return p.sql.Close()
}

// HealthCheck pings the DB using the provided context (with timeout recommended).
func (p *Provider) HealthCheck(ctx context.Context) error {
	if p == nil || p.sql == nil {
		return errors.New("db not initialized")
	}
	return p.sql.PingContext(ctx)
}

// Migrate runs GORM AutoMigrate over provided models.
// Example: provider.Migrate(&User{}, &Order{})
func (p *Provider) Migrate(models ...interface{}) error {
	if p == nil || p.db == nil {
		return errors.New("db not initialized")
	}
	return p.db.AutoMigrate(models...)
}
