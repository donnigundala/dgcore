package http

import (
	"time"

	"github.com/donnigundala/dgcore/config"
)

// Config holds all configuration for the HTTP server.
type Config struct {
	Host        string        `mapstructure:"host"`
	Port        string        `mapstructure:"port"`
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	TLS         TLSConfig     `mapstructure:"tls"`
}

// TLSConfig holds the TLS-specific configuration.
type TLSConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	TLSVersion string `mapstructure:"tls_version"`
}

// init registers the default configuration values with the global config manager.
// This allows the application to have sensible defaults and discoverable settings.
func init() {
	config.Add("server.http", map[string]any{
		"host":         "127.0.0.1",
		"port":         "8080",
		"read_timeout": "5s",
		"write_timeout": "10s",
		"idle_timeout": "120s",
		"tls.enabled":    false,
		"tls.cert_file":  "",
		"tls.key_file":   "",
		"tls.tls_version": "TLS1.2",
	})
}