package http

import "time"

type Config struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	TLS             bool
	TLSVersion      string
	CRTFile         string
	KeyFile         string
}

func defaultConfig() *Config {
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
