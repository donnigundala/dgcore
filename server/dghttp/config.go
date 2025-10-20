package dghttp

import "time"

type Config struct {
	Host            string        `mapstructure:"host" json:"host" yaml:"host" json:"host"`
	Port            string        `mapstructure:"port" json:"port" yaml:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout" yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout" yaml:"shutdown_timeout"`
	TLS             bool          `mapstructure:"tls" json:"tls" yaml:"tls"`
	TLSVersion      string        `mapstructure:"tls_version" json:"tls_version" yaml:"tls_version"`
	CRTFile         string        `mapstructure:"crt_file" json:"crt_file" yaml:"crt_file"`
	KeyFile         string        `mapstructure:"key_file" json:"key_file" yaml:"key_file"`
}

// DefaultConfig returns a HTTP server configuration suitable for local development.
func DefaultConfig() *Config {
	return &Config{
		Host:            "0.0.0.0",
		Port:            "8080",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 15 * time.Second,
		TLS:             false,
		TLSVersion:      "TLSv1.2",
		CRTFile:         "server.crt",
		KeyFile:         "server.key",
	}
}

func SetViperDefaultConfig(prefix *string) map[string]interface{} {
	cfg := DefaultConfig()
	// This function can be used to set default values in a Viper instance if needed.
	p := "http"
	if prefix != nil && *prefix != "" {
		p = *prefix
	}
	p = p + "."

	return map[string]interface{}{
		p + "host":             cfg.Host,
		p + "port":             cfg.Port,
		p + "read_timeout":     cfg.ReadTimeout.String(),
		p + "write_timeout":    cfg.WriteTimeout.String(),
		p + "idle_timeout":     cfg.IdleTimeout.String(),
		p + "shutdown_timeout": cfg.ShutdownTimeout.String(),
		p + "tls":              "false",
		p + "tls_version":      cfg.TLSVersion,
		p + "crt_file":         cfg.CRTFile,
		p + "key_file":         cfg.KeyFile,
	}
}
