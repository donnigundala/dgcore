package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// HTTPServer wraps the standard http.Server to provide a production-ready,
// configurable server that implements the Runnable interface.
type HTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

// HTTPServerOption configures an HTTPServer.
type HTTPServerOption func(*HTTPServer)

// NewHTTPServer creates a new HTTPServer.
// The handler is the root HTTP handler for the server (e.g., a router).
func NewHTTPServer(cfg Config, handler http.Handler, opts ...HTTPServerOption) *HTTPServer {
	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	srv := &HTTPServer{
		server: httpServer,
	}

	for _, opt := range opts {
		opt(srv)
	}

	// Set a default logger if one wasn't provided.
	if srv.logger == nil {
		srv.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	srv.logger = srv.logger.With("component", "http-server", "addr", srv.server.Addr)

	// Apply TLS configuration if enabled.
	if cfg.TLS.Enabled {
		tlsConfig, err := createTLSConfig(cfg.TLS.CertFile, cfg.TLS.KeyFile, cfg.TLS.TLSVersion)
		if err != nil {
			srv.logger.Error("failed to create TLS config", "error", err)
			// Depending on the policy, you might want to panic here.
			// For now, we'll proceed without TLS.
		} else {
			srv.server.TLSConfig = tlsConfig
		}
	}

	return srv
}

// WithHTTPLogger sets a custom logger for the HTTP server.
func WithHTTPLogger(logger *slog.Logger) HTTPServerOption {
	return func(s *HTTPServer) {
		s.logger = logger
	}
}

// WithHTTPHandler sets the HTTP handler for the server.
func WithHTTPHandler(handler http.Handler) HTTPServerOption {
	return func(s *HTTPServer) {
		s.server.Handler = handler
	}
}

// Start begins listening and serving requests. It's a blocking call.
func (s *HTTPServer) Start() error {
	if s.server.TLSConfig != nil {
		s.logger.Info("starting HTTPS server")
		// Cert and key files are now loaded via the TLSConfig.
		return s.server.ListenAndServeTLS("", "")
	}
	s.logger.Info("starting HTTP server")
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.server.Shutdown(ctx)
}

// Addr returns the server's configured address.
func (s *HTTPServer) Addr() string {
	return s.server.Addr
}

// createTLSConfig creates a secure, modern TLS config from file paths.
func createTLSConfig(certFile, keyFile, version string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		return nil, errors.New("TLS certificate or key file not provided")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
	}

	// Map string version to tls constant.
	tlsVersionMap := map[string]uint16{
		"TLS1.2": tls.VersionTLS12,
		"TLS1.3": tls.VersionTLS13,
	}
	minVersion, ok := tlsVersionMap[version]
	if !ok {
		minVersion = tls.VersionTLS12 // Default to a secure modern version.
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   minVersion,
		CipherSuites: []uint16{
			// Modern, secure cipher suites.
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}
