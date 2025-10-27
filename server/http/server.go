package http

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Server struct holds only the final values.
// --- 1. Core Server Struct ---
type Server struct {
	server      *http.Server
	logger      *slog.Logger
	tlsCertFile string
	tlsKeyFile  string
}

// Option The heart of the pattern.
// --- 2. Option Type ---
type Option func(*Server)

// maxHeaderBytes defines the maximum size of HTTP headers (1MB).
const maxHeaderBytes = 1 << 20

func NewServer(handler http.Handler, opts ...Option) *Server {
	srv := &Server{
		server: &http.Server{
			Handler: handler,
			// Set sensible defaults that can be overridden later
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    120 * time.Second,
			MaxHeaderBytes: maxHeaderBytes,
		},
	}

	// Apply all the functional options to override the defaults
	for _, opt := range opts {
		opt(srv)
	}

	// If no logger was provided by the options, create a default one.
	if srv.logger == nil {
		// This default logger writes to standard error, which is a sensible default.
		srv.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// Create a sub-logger for the http-server component.
	srv.logger = srv.logger.With("component", "http-server")

	return srv
}

// --- 4. The Options Themselves ---
// These are the public API for configuring your server.

// WithLogger sets a custom logger for the server.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithAddress sets the host and port for the server.
func WithAddress(host, port string) Option {
	return func(s *Server) {
		s.server.Addr = host + ":" + port
	}
}

// WithTimeouts sets the read, write, and idle timeouts for the server.
func WithTimeouts(read, write, idle time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = read
		s.server.WriteTimeout = write
		s.server.IdleTimeout = idle
	}
}

// WithTLS sets the use of TLS with the given certificate and key files.
func WithTLS(certFile, keyFile, tlsVersion string) Option {
	return func(s *Server) {
		// Use a helper to create a secure, modern TLS config
		s.server.TLSConfig = createTLSConfig(s, tlsVersion)
		s.tlsCertFile = certFile
		s.tlsKeyFile = keyFile
	}
}

// NewServerFromConfig creates a new server from a configuration struct.
// This acts as a bridge between your Viper-based config and the server component.
func NewServerFromConfig(cfg *Config, handler http.Handler) *Server {
	// Create a slice of options from your config struct
	var opts []Option
	opts = append(opts, WithAddress(cfg.Host, cfg.Port))
	opts = append(opts, WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout))

	if cfg.TLS.Enabled {
		opts = append(opts, WithTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile, cfg.TLS.TLSVersion))
	}

	// Call the core constructor with the generated options
	return NewServer(handler, opts...)
}

// Server get http server instance
func (s *Server) Server() *http.Server {
	return s.server
}

// Start http server
func (s *Server) Start() error {
	// Check if TLS is configured by looking for the TLSConfig and cert files
	isTLS := s.server.TLSConfig != nil && s.tlsCertFile != "" && s.tlsKeyFile != ""

	if isTLS {
		s.logger.Info("Starting HTTPS server", "address", s.server.Addr)
		return s.server.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile)
	}

	s.logger.Info("Starting HTTP server", "address", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Close immediately closes all active net.Listeners.
func (s *Server) Close() error {
	return s.server.Close()
}

// Addr returns the server's address
func (s *Server) Addr() string {
	return s.server.Addr
}

// createTLSConfig creates a secure, modern TLS config.
func createTLSConfig(s *Server, version string) *tls.Config {
	// Map string version to tls constant
	tlsVersionMap := map[string]uint16{
		"TLS1.1": tls.VersionTLS11,
		"TLS1.2": tls.VersionTLS12,
		"TLS1.3": tls.VersionTLS13,
	}

	minVersion, ok := tlsVersionMap[version]
	if !ok {
		// Default to a secure modern version
		minVersion = tls.VersionTLS12
		s.logger.Warn(
			"Unsupported TLS version, defaulting to TLS 1.2",
			"unsupported_version", version,
			"default_version", "TLS1.2",
		)
	}

	return &tls.Config{
		MinVersion: minVersion,
		CipherSuites: []uint16{
			// Modern, secure cipher suites
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}
