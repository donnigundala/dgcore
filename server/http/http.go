package http

import (
	"context"
	"crypto/tls"
	"net/http"
)

type Http struct {
	cfg    *Config
	server *http.Server
	ctx    context.Context
}

type IHttp interface {
	Server() *http.Server
	Start() error
	Shutdown() error
	Close() error
	String() string
}

func NewHttp(ctx context.Context, cfg *Config, handler http.Handler) *Http {
	srv := &Http{
		cfg:    cfg,
		server: &http.Server{},
		ctx:    ctx,
	}

	srv.server.Addr = cfg.Host + ":" + cfg.Port
	srv.server.Handler = handler
	srv.server.ReadTimeout = cfg.ReadTimeout
	srv.server.WriteTimeout = cfg.WriteTimeout
	srv.server.IdleTimeout = cfg.IdleTimeout
	srv.server.MaxHeaderBytes = 1 << 20 // 1 MB (default header size limit)

	// Configure TLS if enabled
	if cfg.TLS {
		srv.server.TLSConfig = &tls.Config{
			MinVersion: srv.tlsVersion(),
		}
		srv.server.TLSConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		}
	}

	return srv
}

// Server get http server instance
func (h *Http) Server() *http.Server {
	return h.server
}

// Start http server
func (h *Http) Start() error {
	if h.cfg.TLS {
		logPrintf("[HTTP] Starting HTTPS server on %s:%s", h.cfg.Host, h.cfg.Port)
		return h.server.ListenAndServeTLS(h.cfg.CRTFile, h.cfg.KeyFile)
	} else {
		logPrintf("[HTTP] Starting HTTP server on %s:%s", h.cfg.Host, h.cfg.Port)
		return h.server.ListenAndServe()
	}
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (h *Http) Shutdown() error {
	return h.server.Shutdown(h.ctx)
}

// Close immediately closes all active net.Listeners.
func (h *Http) Close() error {
	return h.server.Close()
}

// String returns the server's address
func (h *Http) String() string {
	return h.server.Addr
}

// tlsVersion maps the TLS version string to the tls package constants
func (h *Http) tlsVersion() uint16 {
	switch h.cfg.TLSVersion {
	case "TLS1.0", "TLSv1.0":
		return tls.VersionTLS10
	case "TLS1.1", "TLSv1.1":
		return tls.VersionTLS11
	case "TLS1.2", "TLSv1.2":
		return tls.VersionTLS12
	case "TLS1.3", "TLSv1.3":
		return tls.VersionTLS13
	default:
		logPrintf(" - TLS version not supported or unknown: %s", h.cfg.TLSVersion)
		logPrintf(" - Falling back to TLS 1.0")
		return tls.VersionTLS10 // Default to TLS 1.0 if nothing is specified
	}
}
