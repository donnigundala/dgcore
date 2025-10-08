package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const (
	defaultMaxIdleConns         = 10
	defaultMaxOpenConns         = 100
	defaultConnMaxLifetime      = 1 * time.Hour
	defaultConnMaxIdleTime      = 30 * time.Minute
	defaultMaxConnectionTimeout = 5 * time.Second
)

// connectionPoolConfig holds the configuration for the database connection pool
type connectionPoolConfig struct {
	MaxIdleConns         int
	MaxOpenConns         int
	ConnMaxLifetime      time.Duration
	ConnMaxIdleTime      time.Duration
	MaxConnectionTimeout time.Duration
}

// nwwConnectionPoolConfig returns a ConnectionPoolConfig with default values
func newConnectionPoolConfig() *connectionPoolConfig {
	return &connectionPoolConfig{
		MaxIdleConns:         defaultMaxIdleConns,
		MaxOpenConns:         defaultMaxOpenConns,
		ConnMaxLifetime:      defaultConnMaxLifetime,
		ConnMaxIdleTime:      defaultConnMaxIdleTime,
		MaxConnectionTimeout: defaultMaxConnectionTimeout,
	}
}

// configureConnectionPool sets up the database connection pool
func configureConnectionPool(db *gorm.DB, cfg *Config) error {
	poolCfg := newConnectionPoolConfig()

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Override defaults with config values if provided
	if cfg.MaxOpenConnection != "" {
		maxOpenConns, err := strconv.Atoi(cfg.MaxOpenConnection)
		if err == nil && maxOpenConns > 0 {
			poolCfg.MaxOpenConns = maxOpenConns
		}
	}

	if cfg.MaxConnectionLifetime != "" {
		if lifetime, err := time.ParseDuration(cfg.MaxConnectionLifetime); err == nil {
			poolCfg.ConnMaxLifetime = lifetime
		}
	}

	if cfg.MaxIdleLifetime != "" {
		if idleTime, err := time.ParseDuration(cfg.MaxIdleLifetime); err == nil {
			poolCfg.ConnMaxIdleTime = idleTime
		}
	}

	// Validate configuration
	if err := validateConnectionPoolConfig(poolCfg); err != nil {
		return fmt.Errorf("invalid connection pool configuration: %w", err)
	}

	sqlDB.SetMaxOpenConns(poolCfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

	// Verify connection pool configuration
	if err := verifyConnectionPool(sqlDB, poolCfg.MaxConnectionTimeout); err != nil {
		return fmt.Errorf("connection pool verification failed: %w", err)
	}

	return nil
}

// validateConnectionPoolConfig validates the connection pool configuration
func validateConnectionPoolConfig(poolCfg *connectionPoolConfig) error {
	if poolCfg.MaxOpenConns <= 0 {
		return fmt.Errorf("MaxOpenConns must be greater than 0")
	}

	if poolCfg.MaxIdleConns > poolCfg.MaxOpenConns {
		return fmt.Errorf("MaxIdleConns (%d) cannot be greater than MaxOpenConns (%d)",
			poolCfg.MaxIdleConns, poolCfg.MaxOpenConns)
	}

	if poolCfg.ConnMaxLifetime <= 0 {
		return fmt.Errorf("ConnMaxLifetime must be greater than 0")
	}

	if poolCfg.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("ConnMaxIdleTime must be greater than 0")
	}

	return nil
}

// verifyConnectionPool performs basic verification of the connection pool settings
func verifyConnectionPool(sqlDB *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to verify connection pool: %w", err)
	}

	stats := sqlDB.Stats()
	log.Printf("[PGSQL] Connection pool configured successfully. MaxOpenConns: %d InUse: %d, Idle: %d",
		stats.MaxOpenConnections,
		stats.InUse,
		stats.Idle,
	)

	return nil
}
