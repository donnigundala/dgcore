package mysql

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	defaultMaxIdleConns         = 10
	defaultMaxOpenConns         = 100
	defaultConnMaxLifetime      = 1 * time.Hour
	defaultConnMaxIdleTime      = 30 * time.Minute
	defaultMaxConnectionTimeout = 5 * time.Second
	defaultMaxRetries           = 3
	defaultRetryDelay           = 2 * time.Second
)

type Connector struct {
	DB     *gorm.DB
	Config *Config
}

func NewMysql(cfg *Config) (*Connector, error) {
	// Validate database configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("[MySQL]: config validation failed: %w", err)
	}

	dsn := CreateDSN(cfg)
	maskedDSN := maskDSBPassword(dsn)
	gormConfig := createGormConfig(cfg)

	// Retry connection logic with exponential backoff
	var db *gorm.DB
	var err error
	var wait time.Duration = defaultRetryDelay

	// Retry connection logic with exponential backoff
	for retries := 0; retries < defaultMaxRetries; retries++ {
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err == nil {
			sqlDB, errPing := db.DB()
			if errPing != nil {
				err = fmt.Errorf("[MySQL]: invalid DB handle: %w", errPing)
			} else if errPing = sqlDB.Ping(); errPing != nil {
				err = fmt.Errorf("[MySQL]: ping failed: %w", errPing)
			} else {
				break
			}
		}

		log.Printf("[MySQL]: Retry %d/%d: failed to connect (%v), retrying in %v...", retries+1, defaultMaxRetries, err, wait)
		time.Sleep(wait)
		wait *= 2 // Exponential backoff
	}

	if err != nil {
		return nil, fmt.Errorf("[MySQL]: failed to connect after %d retries. DSN: (%v) \n\r error: %w", defaultMaxRetries, maskedDSN, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[MySQL]: invalid DB handle: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("[MySQL]: ping failed: %w", err)
	}

	if err := configureConnectionPool(db, cfg); err != nil {
		return nil, fmt.Errorf("[MySQL]: failed to configure pool: %w", err)
	}

	log.Printf("[DATABASE] Connected to %s:%s (%s)", cfg.Host, cfg.Port, cfg.Name)
	return &Connector{DB: db}, nil
}

func (c *Connector) Connect() any {
	return c.DB
}

func (c *Connector) Close() error {
	if c.DB == nil {
		return nil
	}
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("[MYSQL] unable to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
