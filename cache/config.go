package cache

import (
	"log/slog"
	"time"

	"github.com/donnigundala/dgcore/config"
)

// Config defines the top-level configuration for a single cache connection.
type Config struct {
	Driver    Driver        `mapstructure:"driver"`
	Namespace string        `mapstructure:"namespace"`
	Separator string        `mapstructure:"separator"`
	Redis     *RedisConfig  `mapstructure:"redis"`
	Memcache  *MemcacheConfig `mapstructure:"memcache"`
	Logger    *slog.Logger  `mapstructure:"-"`
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
	Servers []string `mapstructure:"servers"`
	TTL     time.Duration `mapstructure:"ttl"`
}

// init registers the default configuration values.
func init() {
	// Registering defaults for a 'default' cache connection.
	// A user can define more, like 'caches.session', 'caches.object', etc.
	config.Add("caches.default", map[string]any{
		"driver":    "redis",
		"namespace": "dgcore",
		"separator": ":",
		"redis": map[string]any{
			"host":      "127.0.0.1",
			"port":      "6379",
			"password":  "",
			"db":        0,
			"ttl":       "1h",
			"enable_tls": false,
		},
		"memcache": map[string]any{
			"servers": []string{"127.0.0.1:11211"},
			"ttl":     "1h",
		},
	})
}