package provider

import "time"

var (
	DriverRedis    Driver = "redis"
	DriverMemcache Driver = "memcache"
)

type Driver string

// Config defines Redis and Memcache connection and caching settings.
type Config struct {
	Driver Driver `mapstructure:"driver" json:"driver" yaml:"driver"` // dgredis

	// Host is the Redis server hostname or IP address.
	Host string `mapstructure:"host" json:"host" yaml:"host"`

	// Port is the Redis server port number.
	Port string `mapstructure:"port" json:"port" yaml:"port"`

	// Username for Redis authentication (if required).
	Username string `mapstructure:"username" json:"username" yaml:"username"`

	// Password for Redis authentication (if required).
	Password string `mapstructure:"password" json:"password" yaml:"password"`

	// DB is the Redis logical database index (0–15).
	// Defaults to 0. Override with REDIS_DB in environment variables.
	DB int `mapstructure:"db" json:"db" yaml:"db"`

	// TTL is the default time-to-live for cached items.
	TTL time.Duration `mapstructure:"ttl" json:"ttl" yaml:"ttl"`

	// EnableTLS indicates whether to use TLS for the Redis connection.
	EnableTLS bool `mapstructure:"enable_tls" json:"enable_tls" yaml:"enable_tls"`

	// Namespace is a prefix added to all keys to avoid collisions.
	Namespace string `mapstructure:"namespace" json:"namespace" yaml:"namespace"`

	// Separator is the string used to separate namespace and keys.
	Separator string `mapstructure:"separator" json:"separator" yaml:"separator"`
}
