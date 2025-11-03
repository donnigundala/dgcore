package dgsql

import (
	"fmt"
	"time"
)

type Config struct {
	Driver Driver `json:"driver" yaml:"driver" mapstructure:"driver"` // mysql, postgres, sqlite

	// Either provide DSN directly or field-level config below
	DSN string `json:"dsn" yaml:"dsn" mapstructure:"dsn"`

	// Field-level config (used if DSN == "")
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     int    `json:"port" yaml:"port" mapstructure:"port"`
	User     string `json:"user" yaml:"user" mapstructure:"user"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	DBName   string `json:"db_name" yaml:"db_name" mapstructure:"db_name"`

	// Extras
	// postgres only (disable, require, verify-ca, verify-full)
	SSLMode string `mapstructure:"ssl_mode" yaml:"ssl_mode" mapstructure:"ssl_mode"`
	// postgres only (e.g., UTC)
	TimeZone string `mapstructure:"timezone" yaml:"timezone" mapstructure:"timezone"`
	// additional query params appended to DSN
	Params map[string]string `mapstructure:"params" yaml:"params" json:"params"`

	// SQLite specifics
	// e.g., file:app.db or :memory:
	SQLitePath string `mapstructure:"sqlite_path" yaml:"sqlite_path" json:"sqlite_path"`

	// Pooling
	MaxOpenConns    int           `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time" json:"conn_max_idle_time"`

	// Connection retries
	// number of retries on initial connect/ping
	MaxConnectRetries int `mapstructure:"max_connect_retries" yaml:"max_connect_retries" json:"max_connect_retries"`
	// base backoff between retries
	ConnectRetryBackoff time.Duration `mapstructure:"connect_retry_backoff" yaml:"connect_retry_backoff" json:"connect_retry_backoff"`

	// GORM options
	SkipDefaultTransaction bool `mapstructure:"skip_default_transaction" yaml:"skip_default_transaction" json:"skip_default_transaction"`
	PrepareStmt            bool `mapstructure:"prepare_stmt" yaml:"prepare_stmt" json:"prepare_stmt"`
	SingularTable          bool `mapstructure:"singular_table" yaml:"singular_table" json:"singular_table"`

	// Logging
	// Log level: silent, error, warn, info
	LogLevel LogLevel `mapstructure:"log_level" yaml:"log_level" json:"log_level"`
	// slow query threshold
	SlowThreshold time.Duration `mapstructure:"slow_threshold" yaml:"slow_threshold" json:"slow_threshold"`

	// Observability
	EnablePrometheus    bool   `mapstructure:"enable_prometheus" yaml:"enable_prometheus" json:"enable_prometheus"`
	PrometheusNamespace string `mapstructure:"prometheus_namespace" yaml:"prometheus_namespace" json:"prometheus_namespace"`
	EnableOTel          bool   `mapstructure:"enable_otel" yaml:"enable_otel" json:"enable_otel"`

	// Startup
	AutoMigrate bool `mapstructure:"auto_migrate" yaml:"auto_migrate" json:"auto_migrate"`

	// Debug mode
	Debug bool `mapstructure:"debug" yaml:"debug" json:"debug"`
}

// DSNFor builds a DSN string for the configured driver.
// If Config.DSN is non-empty, it's returned as-is.
func (c *Config) DSNFor() (string, error) {
	if c.DSN != "" {
		return c.DSN, nil
	}
	switch c.Driver {
	case DriverMySQL:
		// Example:
		// user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=UTC&timeout=5s&readTimeout=30s&writeTimeout=30s
		base := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.User, c.Password, c.Host, c.Port, c.DBName)
		params := map[string]string{
			"charset":      "utf8mb4",
			"parseTime":    "True",
			"loc":          "UTC",
			"timeout":      "5s",
			"readTimeout":  "30s",
			"writeTimeout": "30s",
		}
		for k, v := range c.Params {
			params[k] = v
		}
		return base + "?" + encodeParams(params), nil

	case DriverPostgres:
		// Example:
		// host=... user=... password=... dbname=... port=... sslmode=disable TimeZone=UTC
		base := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode, c.TimeZone,
		)
		// Additional params (if present) appended as key='value'
		if len(c.Params) == 0 {
			return base, nil
		}
		extra := ""
		for k, v := range c.Params {
			extra += fmt.Sprintf(" %s=%q", k, v)
		}
		return base + extra, nil

	case DriverSQLite:
		// Example:
		// file:app.db?cache=shared&_fk=1
		params := map[string]string{
			"cache": "shared",
			"_fk":   "1",
		}
		for k, v := range c.Params {
			params[k] = v
		}
		return c.SQLitePath + "?" + encodeParams(params), nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", c.Driver)
	}
}
