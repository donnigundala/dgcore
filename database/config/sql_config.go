package config

import "time"

// LogLevel represents the verbosity of GORM's logger.
type LogLevel string

const (
	LogLevelSilent LogLevel = "silent"
	LogLevelError  LogLevel = "error"
	LogLevelWarn   LogLevel = "warn"
	LogLevelInfo   LogLevel = "info"
)

// SQLTLSConfig contains the TLS settings for a SQL database.
// For GORM, TLS is often configured via DSN parameters or specific driver options.
// This struct provides a generic way to enable/disable TLS and specify certs.
// The actual application of these settings will depend on the specific GORM driver.
type SQLTLSConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	CA       Secret `json:"ca" mapstructure:"ca"`     // CA certificate content
	Cert     Secret `json:"cert" mapstructure:"cert"`   // Client certificate content
	Key      Secret `json:"key" mapstructure:"key"`    // Client key content
	Insecure bool   `json:"insecure" mapstructure:"insecure"` // Skip TLS verification
}

// SQLPoolConfig contains the connection pool settings for a SQL database.
type SQLPoolConfig struct {
	MaxOpenConns    int           `json:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
}

// SQLConnectionDetails contains individual parameters for a SQL database connection.
type SQLConnectionDetails struct {
	Host     Secret            `json:"host" mapstructure:"host"`
	Port     Secret            `json:"port" mapstructure:"port"`
	User     Secret            `json:"user" mapstructure:"user"`
	Password Secret            `json:"password" mapstructure:"password"`
	DBName   Secret            `json:"db_name" mapstructure:"db_name"`
	Params   map[string]Secret `json:"params" mapstructure:"params"` // Additional driver-specific parameters
	RawDSN   Secret            `json:"raw_dsn" mapstructure:"raw_dsn"`   // Fallback for direct DSN string
}

type SQLConfig struct {
	DriverName string                `json:"driver_name" mapstructure:"driver_name"` // e.g., "postgres", "mysql", "sqlite"
	Primary    *SQLConnectionDetails `json:"primary" mapstructure:"primary"`
	Replicas   []*SQLConnectionDetails `json:"replicas" mapstructure:"replicas"`
	Pool       *SQLPoolConfig        `json:"pool" mapstructure:"pool"`
	TLS        *SQLTLSConfig         `json:"tls" mapstructure:"tls"`
	LogLevel   LogLevel              `json:"log_level" mapstructure:"log_level"`
}
