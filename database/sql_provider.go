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
)

// sqlProvider implements the SQLProvider interface using GORM.
type sqlProvider struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSQLProvider creates a new GORM-based SQL provider.
func NewSQLProvider(ctx context.Context, cfg *SQLConfig, policy *PolicyConfig, logger *slog.Logger) (Provider, error) {
	dsn := buildDSN(cfg)
	if dsn == "" {
		return nil, fmt.Errorf("could not build DSN for driver: %s", cfg.DriverName)
	}

	var dialector gorm.Dialector
	switch cfg.DriverName {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported SQL driver: %s", cfg.DriverName)
	}

	// Configure GORM logger
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

	if err := provider.Ping(ctx); err != nil {
		return nil, fmt.Errorf("initial database ping failed: %w", err)
	}

	logger.Info("SQL database connection successful")
	return provider, nil
}

func (p *sqlProvider) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (p *sqlProvider) Close() error {
	p.logger.Info("Closing SQL database connection...")
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (p *sqlProvider) Gorm() interface{} {
	return p.db
}

// buildDSN constructs the Data Source Name from the config.
// It no longer resolves env vars itself; that is handled by Viper.
func buildDSN(cfg *SQLConfig) string {
	if cfg.DSN != nil && cfg.DSN.Primary != "" {
		return cfg.DSN.Primary
	}

	p := cfg.Primary
	switch cfg.DriverName {
	case "postgres":
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

var gormLogMapping = map[string]gormlogger.LogLevel{
	"silent": gormlogger.Silent,
	"error":  gormlogger.Error,
	"warn":   gormlogger.Warn,
	"info":   gormlogger.Info,
}
