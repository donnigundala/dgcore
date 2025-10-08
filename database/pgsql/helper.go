package pgsql

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// validateDBConfig validates the database configuration
func validateDBConfig(cfg *Config) error {
	required := map[string]string{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"name":     cfg.Name,
		"username": cfg.Username,
	}

	for field, value := range required {
		if value == "" {
			return fmt.Errorf("[PGSQL] database %s is required", field)
		}
	}

	// Validate numeric values
	if port, err := strconv.Atoi(cfg.Port); err != nil || port <= 0 {
		return fmt.Errorf("[PGSQL] invalid port number")
	}

	return nil
}

// PostgresDSN creates the database connection string
func PostgresDSN(cfg *Config) string {
	return fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s TimeZone=%s user=%s password=%s",
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.Timezone,
		cfg.Username,
		cfg.Password,
	)
}

// createGormConfig creates the GORM configuration
func createGormConfig(cfg *Config) *gorm.Config {
	gormConfig := &gorm.Config{}
	if cfg.Debug {
		gormConfig.Logger = gormLogger.Default.LogMode(gormLogger.Info)
	}
	return gormConfig
}

// maskDSBPassword hides the password in the DSN string for logging
func maskDSBPassword(dsn string) string {
	// This is a simple implementation assuming password= is present and followed by a space or end of string
	const pwKey = "password="
	start := -1
	for i := 0; i+len(pwKey) <= len(dsn); i++ {
		if dsn[i:i+len(pwKey)] == pwKey {
			start = i + len(pwKey)
			break
		}
	}
	if start == -1 {
		return dsn
	}
	end := start
	for end < len(dsn) && dsn[end] != ' ' {
		end++
	}
	return dsn[:start] + "****" + dsn[end:]
}
