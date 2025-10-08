package pgsql

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultMaxRetries = 3
	defaultRetryDelay = 2 * time.Second
)

type Connector struct {
	DB     *gorm.DB
	Config *Config
}

// NewPostgres creates a new PostgresSQL database connection using the provided configuration.
// It includes connection retry logic and proper error handling.
func NewPostgres(cfg *Config) (*Connector, error) {
	// Validate database configuration
	if err := validateDBConfig(cfg); err != nil {
		return nil, fmt.Errorf("[PGSQL] config validation failed: %w", err)
	}

	dsn := PostgresDSN(cfg)
	maskedDSN := maskDSBPassword(dsn)
	gormConfig := createGormConfig(cfg)

	// Retry connection logic with exponential backoff
	var db *gorm.DB
	var err error
	var wait time.Duration = defaultRetryDelay

	// Retry connection logic with exponential backoff
	for retries := 0; retries < defaultMaxRetries; retries++ {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			sqlDB, errPing := db.DB()
			if errPing != nil {
				err = fmt.Errorf("[PGSQL] invalid DB handle: %w", errPing)
			} else if errPing = sqlDB.Ping(); errPing != nil {
				err = fmt.Errorf("[PGSQL] ping failed: %w", errPing)
			} else {
				break
			}
		}

		log.Printf("[PGSQL] Retry %d/%d: failed to connect (%v), retrying in %v...", retries+1, defaultMaxRetries, err, wait)
		time.Sleep(wait)
		wait *= 2 // Exponential backoff
	}

	if err != nil {
		return nil, fmt.Errorf("[PGSQL] failed to connect after %d retries. DSN: (%v) \n\r error: %w", defaultMaxRetries, maskedDSN, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[PGSQL] invalid DB handle: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("[PGSQL] ping failed: %w", err)
	}

	if err := configureConnectionPool(db, cfg); err != nil {
		return nil, fmt.Errorf("[PGSQL] failed to configure pool: %w", err)
	}

	log.Printf("[PGSQL] Connected to %s:%s (db:%s)", cfg.Host, cfg.Port, cfg.Name)
	//return db, nil
	return &Connector{DB: db, Config: cfg}, nil
}

func (c *Connector) Connect() (any, error) {
	return c.DB, nil
}

func (c *Connector) Close() error {
	if c.DB == nil {
		return nil
	}
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("[PGSQL] unable to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
