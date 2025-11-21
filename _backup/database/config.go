package database

import "time"

// --- Top Level Config ---

// Config is the root configuration for the entire database manager.
type Config struct {
	DefaultConnection string                `mapstructure:"default_connection"`
	Connections       map[string]Connection `mapstructure:"connections"`
}

// Connection is the configuration for a single, named database connection.
type Connection struct {
	Driver string       `mapstructure:"driver"`
	Policy PolicyConfig `mapstructure:"policy"`
	SQL    *SQLConfig   `mapstructure:"sql"`
	Mongo  *MongoConfig `mapstructure:"mongo"`
}

// --- Policy ---

// PolicyConfig defines the reconnection and health check policy for a connection.
type PolicyConfig struct {
	PingInterval         time.Duration `mapstructure:"ping_interval"`
	MaxFailures          int           `mapstructure:"max_failures"`
	ReconnectBackoffBase time.Duration `mapstructure:"reconnect_backoff_base"`
	ReconnectBackoffMax  time.Duration `mapstructure:"reconnect_backoff_max"`
}

// --- SQL ---

// SQLConfig holds all settings for a SQL-based connection.
type SQLConfig struct {
	DriverName string       `mapstructure:"driver_name"`
	Primary    NodeConfig   `mapstructure:"primary"`
	Replicas   []NodeConfig `mapstructure:"replicas"`
	Pool       PoolConfig   `mapstructure:"pool"`
	TLS        TLSConfig    `mapstructure:"tls"`
	LogLevel   string       `mapstructure:"log_level"`
	DSN        *DSNConfig   `mapstructure:"dsn"`
}

// NodeConfig represents a single database node.
type NodeConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"` // Changed from Secret to string
	DBName   string `mapstructure:"db_name"`
}

// PoolConfig defines connection pool settings for SQL.
type PoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// TLSConfig defines TLS settings for SQL.
type TLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	CertPath           string `mapstructure:"cert_path"`
	KeyPath            string `mapstructure:"key_path"`
	CAPath             string `mapstructure:"ca_path"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

// DSNConfig allows providing a raw DSN string.
type DSNConfig struct {
	Primary  string   `mapstructure:"primary"`  // Changed from Secret to string
	Replicas []string `mapstructure:"replicas"` // Changed from Secret to string
}

// --- MongoDB ---

// MongoConfig holds all settings for a MongoDB connection.
type MongoConfig struct {
	URI      string    `mapstructure:"uri"` // Changed from Secret to string
	Database string    `mapstructure:"database"`
	Pool     MongoPool `mapstructure:"pool"`
	LogLevel string    `mapstructure:"log_level"`
}

// MongoPool holds MongoDB connection pool settings.
type MongoPool struct {
	MaxPoolSize uint64 `mapstructure:"max_pool_size"`
}
