package mysql

import (
	"fmt"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// CreateDSN creates the database connection string
func CreateDSN(cfg *Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}

func validateConfig(cfg *Config) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.Username == "" || cfg.Name == "" {
		return fmt.Errorf("missing required SQL config field")
	}
	return nil
}

func createGormConfig(cfg *Config) *gorm.Config {
	gormCfg := &gorm.Config{}
	if cfg.Debug {
		gormCfg.Logger = gormLogger.Default.LogMode(gormLogger.Info)
	}
	return gormCfg
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
