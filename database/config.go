// Package database provides a database provider/factory using GORM
// for MySQL, PostgreSQL, and SQLite with best-practice defaults.
package database

import (
	"fmt"
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
	Driver Driver `json:"driver" yaml:"driver" mapstructure:"driver"` // mysql, postgres, sqlite

	// Either provide DSN directly or field-level config below
	DSN string `json:"dsn" yaml:"dsn" mapstructure:"dsn"`

	// Field-level config (used if DSN == "")
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     int    `json:"port" yaml:"port" mapstructure:"port"`
	User     string `json:"user" yaml:"user" mapstructure:"user"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	DBName   string `json:"db_name" yaml:"db_name" mapstructure:"dbname"`

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

// DefaultConfig applies sensible defaults if fields are zero-valued.
func DefaultConfig() *Config {
	c := &Config{}
	if c.Driver == "" {
		c.Driver = DriverPostgres
	}
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		switch c.Driver {
		case DriverMySQL:
			c.Port = 3306
		case DriverPostgres:
			c.Port = 5432
		case DriverSQLite:
			c.Port = 0 // not used
		default:
			c.Port = 5432 // default to Postgres port
		}
	}
	if c.User == "" {
		switch c.Driver {
		case DriverMySQL:
			c.User = "root"
		case DriverPostgres:
			c.User = "postgres"
		case DriverSQLite:
			c.User = "" // not used
		default:
			c.User = "postgres" // default to Postgres user
		}
	}
	if c.DBName == "" {
		c.DBName = "app"
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

	return c
}

//func LoadViperDefault(v *viper.Viper, prefix *string) {
//	cfg := DefaultConfig()
//	// Default prefix
//	p := "db"
//	if prefix != nil && *prefix != "" {
//		p = *prefix
//	}
//	p = p + "."
//
//	v.SetDefault(p+"driver", cfg.Driver)
//	v.SetDefault(p+"dsn", cfg.DSN)
//	v.SetDefault(p+"host", cfg.Host)
//	v.SetDefault(p+"user", cfg.User)
//	v.SetDefault(p+"password", cfg.Password)
//	v.SetDefault(p+"db_name", cfg.DBName)
//	v.SetDefault(p+"ssl_mode", cfg.SSLMode)
//	v.SetDefault(p+"timezone", cfg.TimeZone)
//	v.SetDefault(p+"sqlite_path", cfg.SQLitePath)
//	v.SetDefault(p+"max_open_conns", cfg.MaxOpenConns)
//	v.SetDefault(p+"max_idle_conns", cfg.MaxIdleConns)
//	v.SetDefault(p+"conn_max_lifetime", cfg.ConnMaxLifetime)
//	v.SetDefault(p+"conn_max_idle_time", cfg.ConnMaxIdleTime)
//	v.SetDefault(p+"max_connect_retries", cfg.MaxConnectRetries)
//	v.SetDefault(p+"connect_retry_backoff", cfg.ConnectRetryBackoff)
//	v.SetDefault(p+"skip_default_transaction", cfg.SkipDefaultTransaction)
//	v.SetDefault(p+"prepare_stmt", cfg.PrepareStmt)
//	v.SetDefault(p+"singular_table", cfg.SingularTable)
//	v.SetDefault(p+"log_level", cfg.LogLevel)
//	v.SetDefault(p+"slow_threshold", cfg.SlowThreshold)
//	v.SetDefault(p+"enable_prometheus", cfg.EnablePrometheus)
//	v.SetDefault(p+"prometheus_namespace", cfg.PrometheusNamespace)
//	v.SetDefault(p+"enable_otel", cfg.EnableOTel)
//	v.SetDefault(p+"auto_migrate", cfg.AutoMigrate)
//	v.SetDefault(p+"debug", cfg.Debug)
//}

func SetViperDefaultConfig(prefix *string) map[string]interface{} {
	// This function can be used to set default values in a Viper instance if needed.
	// Uses same defaults as DefaultConfig()
	// Note: prefix should be uppercase if used with env vars
	// e.g., "DB" to set DB_DRIVER, DB_DSN, etc.
	// If prefix is nil or empty, no prefix is used.
	// Example: "db.driver", "db.dsn", etc.
	// If prefix is "MYAPP", then "MYAPP.driver", "MYAPP.dsn", etc.
	// This allows multiple DB configs in same app if needed.

	cfg := DefaultConfig()
	// Default prefix
	p := "db"
	if prefix != nil && *prefix != "" {
		p = *prefix
	}
	p = p + "."
	return map[string]interface{}{
		p + "driver":                   cfg.Driver,
		p + "dsn":                      cfg.DSN,
		p + "host":                     cfg.Host,
		p + "port":                     cfg.Port,
		p + "user":                     cfg.User,
		p + "password":                 cfg.Password,
		p + "db_name":                  cfg.DBName,
		p + "ssl_mode":                 cfg.SSLMode,
		p + "timezone":                 cfg.TimeZone,
		p + "sqlite_path":              cfg.SQLitePath,
		p + "max_open_conns":           cfg.MaxOpenConns,
		p + "max_idle_conns":           cfg.MaxIdleConns,
		p + "conn_max_lifetime":        cfg.ConnMaxLifetime,
		p + "conn_max_idle_time":       cfg.ConnMaxIdleTime,
		p + "max_connect_retries":      cfg.MaxConnectRetries,
		p + "connect_retry_backoff":    cfg.ConnectRetryBackoff,
		p + "skip_default_transaction": cfg.SkipDefaultTransaction,
		p + "prepare_stmt":             cfg.PrepareStmt,
		p + "singular_table":           cfg.SingularTable,
		p + "log_level":                cfg.LogLevel,
		p + "slow_threshold":           cfg.SlowThreshold,
		p + "enable_prometheus":        cfg.EnablePrometheus,
		p + "prometheus_namespace":     cfg.PrometheusNamespace,
		p + "enable_otel":              cfg.EnableOTel,
		p + "auto_migrate":             cfg.AutoMigrate,
		p + "debug":                    cfg.Debug,
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
//func LoadFromEnv(prefix string) (Config, error) {
//	get := func(key, def string) string {
//		if v, ok := os.LookupEnv(key); ok {
//			return v
//		}
//		return def
//	}
//	toInt := func(s string, def int) int {
//		if s == "" {
//			return def
//		}
//		if i, err := strconv.Atoi(s); err == nil {
//			return i
//		}
//		return def
//	}
//	toDur := func(s string, def time.Duration) time.Duration {
//		if s == "" {
//			return def
//		}
//		if d, err := time.ParseDuration(s); err == nil {
//			return d
//		}
//		return def
//	}
//
//	cfg := Config{
//		Driver:          Driver(get(prefix+"_DRIVER", "postgres")),
//		DSN:             get(prefix+"_DSN", ""),
//		Host:            get(prefix+"_HOST", "localhost"),
//		Port:            toInt(get(prefix+"_PORT", "5432"), 5432),
//		User:            get(prefix+"_USER", "postgres"),
//		Password:        get(prefix+"_PASSWORD", ""),
//		DBName:          get(prefix+"_NAME", "app"),
//		SSLMode:         get(prefix+"_SSLMODE", "disable"),
//		TimeZone:        get(prefix+"_TIMEZONE", "UTC"),
//		SQLitePath:      get(prefix+"_SQLITE_PATH", "file:app.db"),
//		MaxOpenConns:    toInt(get(prefix+"_MAX_OPEN_CONNS", "25"), 25),
//		MaxIdleConns:    toInt(get(prefix+"_MAX_IDLE_CONNS", "5"), 5),
//		ConnMaxLifetime: toDur(get(prefix+"_CONN_MAX_LIFETIME", "1h"), time.Hour),
//		ConnMaxIdleTime: toDur(get(prefix+"_CONN_MAX_IDLE_TIME", "30m"), 30*time.Minute),
//
//		// New: connect retry/backoff
//		MaxConnectRetries:   toInt(get(prefix+"_MAX_CONNECT_RETRIES", "5"), 5),
//		ConnectRetryBackoff: toDur(get(prefix+"_CONNECT_RETRY_BACKOFF", "500ms"), 500*time.Millisecond),
//
//		SkipDefaultTransaction: get(prefix+"_SKIP_DEFAULT_TX", "true") == "true",
//		PrepareStmt:            get(prefix+"_PREPARE_STMT", "false") == "true",
//		SingularTable:          get(prefix+"_SINGULAR_TABLE", "false") == "true",
//		LogLevel:               LogLevel(get(prefix+"_LOG_LEVEL", "warn")),
//		SlowThreshold:          toDur(get(prefix+"_SLOW_THRESHOLD", "200ms"), 200*time.Millisecond),
//
//		// New: observability
//		EnablePrometheus:    get(prefix+"_ENABLE_PROMETHEUS", "false") == "true",
//		PrometheusNamespace: get(prefix+"_PROM_NAMESPACE", "gorm"),
//		EnableOTel:          get(prefix+"_ENABLE_OTEL", "false") == "true",
//
//		AutoMigrate: get(prefix+"_AUTO_MIGRATE", "false") == "true",
//	}
//	cfg.Defaults()
//	return cfg, nil
//}

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
