// Package db provides a database provider/factory using GORM
// for MySQL, PostgreSQL, and SQLite with best-practice defaults.
package database

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Driver string

const (
	DriverMySQL    Driver = "mysql"
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite"
)

type LogLevel string

const (
	LogSilent LogLevel = "silent"
	LogError  LogLevel = "error"
	LogWarn   LogLevel = "warn"
	LogInfo   LogLevel = "info"
)

type Config struct {
	Driver Driver

	// Either provide DSN directly or field-level config below
	DSN string

	// Field-level config (used if DSN == "")
	Host     string
	Port     int
	User     string
	Password string
	DBName   string

	// Extras
	SSLMode  string            // postgres only (disable, require, verify-ca, verify-full)
	TimeZone string            // postgres only (e.g., UTC)
	Params   map[string]string // additional query params appended to DSN

	// SQLite specifics
	SQLitePath string // e.g., file:app.db or :memory:

	// Pooling
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Connection retries
	MaxConnectRetries   int           // number of retries on initial connect/ping
	ConnectRetryBackoff time.Duration // base backoff between retries

	// GORM options
	SkipDefaultTransaction bool
	PrepareStmt            bool
	SingularTable          bool

	// Logging
	LogLevel      LogLevel
	SlowThreshold time.Duration // slow query threshold

	// Observability
	EnablePrometheus    bool
	PrometheusNamespace string
	EnableOTel          bool

	// Startup
	AutoMigrate bool
}

// Defaults applies sensible defaults if fields are zero-valued.
func (c *Config) Defaults() {
	if c.Driver == "" {
		c.Driver = DriverPostgres
	}
	if c.Params == nil {
		c.Params = map[string]string{}
	}
	// Pool defaults
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 25
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 5
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = time.Hour
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 30 * time.Minute
	}
	// Logging defaults
	if c.LogLevel == "" {
		c.LogLevel = LogWarn
	}
	if c.SlowThreshold == 0 {
		c.SlowThreshold = 200 * time.Millisecond
	}
	// Connection retry defaults
	if c.MaxConnectRetries == 0 {
		c.MaxConnectRetries = 5
	}
	if c.ConnectRetryBackoff == 0 {
		c.ConnectRetryBackoff = 500 * time.Millisecond
	}
	// Observability defaults
	if c.PrometheusNamespace == "" {
		c.PrometheusNamespace = "gorm"
	}
	// Postgres defaults
	if c.Driver == DriverPostgres {
		if c.SSLMode == "" {
			c.SSLMode = "disable"
		}
		if c.TimeZone == "" {
			c.TimeZone = "UTC"
		}
	}
	// SQLite defaults
	if c.Driver == DriverSQLite && c.SQLitePath == "" {
		c.SQLitePath = "file:app.db"
	}
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

// LoadFromEnv loads configuration from environment variables using a prefix.
// For example, prefix "DB" reads DB_DRIVER, DB_DSN, DB_HOST, DB_PORT, etc.
func LoadFromEnv(prefix string) (Config, error) {
	get := func(key, def string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return def
	}
	toInt := func(s string, def int) int {
		if s == "" {
			return def
		}
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
		return def
	}
	toDur := func(s string, def time.Duration) time.Duration {
		if s == "" {
			return def
		}
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
		return def
	}

	cfg := Config{
		Driver:          Driver(get(prefix+"_DRIVER", "postgres")),
		DSN:             get(prefix+"_DSN", ""),
		Host:            get(prefix+"_HOST", "localhost"),
		Port:            toInt(get(prefix+"_PORT", "5432"), 5432),
		User:            get(prefix+"_USER", "postgres"),
		Password:        get(prefix+"_PASSWORD", ""),
		DBName:          get(prefix+"_NAME", "app"),
		SSLMode:         get(prefix+"_SSLMODE", "disable"),
		TimeZone:        get(prefix+"_TIMEZONE", "UTC"),
		SQLitePath:      get(prefix+"_SQLITE_PATH", "file:app.db"),
		MaxOpenConns:    toInt(get(prefix+"_MAX_OPEN_CONNS", "25"), 25),
		MaxIdleConns:    toInt(get(prefix+"_MAX_IDLE_CONNS", "5"), 5),
		ConnMaxLifetime: toDur(get(prefix+"_CONN_MAX_LIFETIME", "1h"), time.Hour),
		ConnMaxIdleTime: toDur(get(prefix+"_CONN_MAX_IDLE_TIME", "30m"), 30*time.Minute),

		// New: connect retry/backoff
		MaxConnectRetries:   toInt(get(prefix+"_MAX_CONNECT_RETRIES", "5"), 5),
		ConnectRetryBackoff: toDur(get(prefix+"_CONNECT_RETRY_BACKOFF", "500ms"), 500*time.Millisecond),

		SkipDefaultTransaction: get(prefix+"_SKIP_DEFAULT_TX", "true") == "true",
		PrepareStmt:            get(prefix+"_PREPARE_STMT", "false") == "true",
		SingularTable:          get(prefix+"_SINGULAR_TABLE", "false") == "true",
		LogLevel:               LogLevel(get(prefix+"_LOG_LEVEL", "warn")),
		SlowThreshold:          toDur(get(prefix+"_SLOW_THRESHOLD", "200ms"), 200*time.Millisecond),

		// New: observability
		EnablePrometheus:    get(prefix+"_ENABLE_PROMETHEUS", "false") == "true",
		PrometheusNamespace: get(prefix+"_PROM_NAMESPACE", "gorm"),
		EnableOTel:          get(prefix+"_ENABLE_OTEL", "false") == "true",

		AutoMigrate: get(prefix+"_AUTO_MIGRATE", "false") == "true",
	}
	cfg.Defaults()
	return cfg, nil
}

func encodeParams(params map[string]string) string {
	first := true
	out := ""
	for k, v := range params {
		if first {
			out += fmt.Sprintf("%s=%s", k, v)
			first = false
			continue
		}
		out += fmt.Sprintf("&%s=%s", k, v)
	}
	return out
}
