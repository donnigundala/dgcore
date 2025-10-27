package http

import "time"

type Config struct {
	Host         string        `mapstructure:"host" json:"host" yaml:"host" json:"host"`
	Port         string        `mapstructure:"port" json:"port" yaml:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" json:"idle_timeout" yaml:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls" yaml:"tls" json:"tls"`
	//ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout" yaml:"shutdown_timeout"`
	//TLS             bool          `mapstructure:"tls" json:"tls" yaml:"tls"`
	//TLSVersion      string        `mapstructure:"tls_version" json:"tls_version" yaml:"tls_version"`
	//CRTFile         string        `mapstructure:"crt_file" json:"crt_file" yaml:"crt_file"`
	//KeyFile         string        `mapstructure:"key_file" json:"key_file" yaml:"key_file"`
}

// TLSConfig allows enabling HTTPS/TLS easily.
type TLSConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	TLSVersion string `mapstructure:"tls_version" json:"tls_version" yaml:"tls_version"`
	CertFile   string `mapstructure:"cert_file" yaml:"cert_file" json:"cert_file"`
	KeyFile    string `mapstructure:"key_file" yaml:"key_file" json:"key_file"`
}
