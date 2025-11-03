package config

import "time"

// MongoTLSConfig contains the TLS settings for a MongoDB database.
type MongoTLSConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	CAFile   Secret `json:"ca_file" mapstructure:"ca_file"`
	CertFile Secret `json:"cert_file" mapstructure:"cert_file"`
	KeyFile  Secret `json:"key_file" mapstructure:"key_file"`
}

// MongoPoolConfig contains the connection pool settings for a MongoDB database.
type MongoPoolConfig struct {
	MaxPoolSize     uint64        `json:"max_pool_size" mapstructure:"max_pool_size"`
	MinPoolSize     uint64        `json:"min_pool_size" mapstructure:"min_pool_size"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time" mapstructure:"max_conn_idle_time"`
}

type MongoConfig struct {
	PrimaryURI Secret   `json:"primary_uri" mapstructure:"primary_uri"`
	Replicas   []Secret `json:"replicas" mapstructure:"replicas"`
	Database   string   `json:"database" mapstructure:"database"`
	Pool       *MongoPoolConfig `json:"pool" mapstructure:"pool"`
	TLS        *MongoTLSConfig  `json:"tls" mapstructure:"tls"`
	LogLevel   LogLevel         `json:"log_level" mapstructure:"log_level"`
}
