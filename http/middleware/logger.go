package middleware

import (
	"net/http"
	"time"

	"github.com/donnigundala/dg-core/logging"
)

// LoggerConfig defines the configuration for the logger middleware
type LoggerConfig struct {
	Logger        *logging.Logger
	SkipPaths     []string
	LogLatency    bool
	LogClientIP   bool
	LogMethod     bool
	LogPath       bool
	LogStatusCode bool
}

// DefaultLoggerConfig returns the default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Logger:        logging.Default(),
		SkipPaths:     []string{},
		LogLatency:    true,
		LogClientIP:   true,
		LogMethod:     true,
		LogPath:       true,
		LogStatusCode: true,
	}
}

// Logger returns a middleware that logs HTTP requests
func Logger(config LoggerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for certain paths
			for _, path := range config.SkipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			start := time.Now()
			path := r.URL.Path
			raw := r.URL.RawQuery

			// Create a custom response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Process request
			next.ServeHTTP(ww, r)

			// Build log attributes
			attrs := make([]interface{}, 0, 10)

			if config.LogStatusCode {
				attrs = append(attrs, "status", ww.status)
			}

			if config.LogMethod {
				attrs = append(attrs, "method", r.Method)
			}

			if config.LogPath {
				fullPath := path
				if raw != "" {
					fullPath = path + "?" + raw
				}
				attrs = append(attrs, "path", fullPath)
			}

			if config.LogClientIP {
				clientIP := r.RemoteAddr
				if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
					clientIP = forwarded
				} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
					clientIP = realIP
				}
				attrs = append(attrs, "ip", clientIP)
			}

			if config.LogLatency {
				latency := time.Since(start)
				attrs = append(attrs, "latency", latency)
			}

			// Log the request
			config.Logger.Info("Request", attrs...)
		})
	}
}

// LoggerWithDefault returns a logger middleware with default configuration
func LoggerWithDefault() func(http.Handler) http.Handler {
	return Logger(DefaultLoggerConfig())
}

// responseWriter is a wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
