package cache

import (
	"log/slog"
	"time"
)

// Config defines the top-level configuration for a single cache connection.
// This struct is now intended to be filled by the consumer app from a config file.
type Config struct {
	Driver    Driver          `mapstructure:"driver"`
	Namespace string          `mapstructure:"namespace"`
	Separator string          `mapstructure:"separator"`
	Redis     *RedisConfig    `mapstructure:"redis"`
	Memcache  *MemcacheConfig `mapstructure:"memcache"`
	Logger    *slog.Logger    `mapstructure:"-"` // Logger is passed programmatically, not from config files.
}

// RedisConfig holds Redis-specific settings.
type RedisConfig struct {
	Host               string        `mapstructure:"host"`
	Port               string        `mapstructure:"port"`
	Username           string        `mapstructure:"username"`
	Password           string        `mapstructure:"password"`
	DB                 int           `mapstructure:"db"`
	TTL                time.Duration `mapstructure:"ttl"`
	EnableTLS          bool          `mapstructure:"enable_tls"`
	InsecureSkipVerify bool          `mapstructure:"insecure_skip_verify"`
}

// MemcacheConfig holds Memcache-specific settings.
type MemcacheConfig struct {
	Servers []string      `mapstructure:"servers"`
	TTL     time.Duration `mapstructure:"ttl"`
}
