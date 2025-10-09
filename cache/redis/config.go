package redis

import "time"

// -----------------------------
// Config
// -----------------------------

// Config defines Redis connection and caching settings.
type Config struct {
	Host      string        `mapstructure:"host" json:"host" yaml:"host"`
	Port      string        `mapstructure:"port" json:"port" yaml:"port"`
	Username  string        `mapstructure:"username" json:"username" yaml:"username"`
	Password  string        `mapstructure:"password" json:"password" yaml:"password"`
	DB        int           `mapstructure:"db" json:"db" yaml:"db"`
	TTL       time.Duration `mapstructure:"ttl" json:"ttl" yaml:"ttl"`
	EnableTLS bool          `mapstructure:"enable_tls" json:"enable_tls" yaml:"enable_tls"`
	Namespace string        `mapstructure:"namespace" json:"namespace" yaml:"namespace"`
	Separator string        `mapstructure:"separator" json:"separator" yaml:"separator"`
}

// DefaultConfig returns a Redis configuration suitable for local development.
func DefaultConfig() *Config {
	return &Config{
		Host:      "localhost",
		Port:      "6379",
		Username:  "",
		Password:  "",
		DB:        0,
		TTL:       5 * time.Minute,
		EnableTLS: false,
		Namespace: "cache",
		Separator: ":",
	}
}
