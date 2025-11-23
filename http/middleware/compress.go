package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressConfig defines the configuration for compression middleware.
type CompressConfig struct {
	// Level is the compression level (0-9).
	// Default: gzip.DefaultCompression
	Level int

	// MinSize is the minimum response size to compress (in bytes).
	// Default: 1024 (1KB)
	MinSize int

	// ContentTypes is a list of content types to compress.
	// Default: ["text/", "application/json", "application/javascript", "application/xml"]
	ContentTypes []string
}

// DefaultCompressConfig returns the default compression configuration.
func DefaultCompressConfig() CompressConfig {
	return CompressConfig{
		Level:   gzip.DefaultCompression,
		MinSize: 1024,
		ContentTypes: []string{
			"text/",
			"application/json",
			"application/javascript",
			"application/xml",
		},
	}
}

// Compress returns a middleware that compresses HTTP responses.
func Compress(config CompressConfig) func(http.Handler) http.Handler {
	// Apply defaults
	if config.Level == 0 {
		config.Level = gzip.DefaultCompression
	}
	if config.MinSize == 0 {
		config.MinSize = 1024
	}
	if len(config.ContentTypes) == 0 {
		config.ContentTypes = []string{
			"text/",
			"application/json",
			"application/javascript",
			"application/xml",
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Create gzip writer
			gz, err := gzip.NewWriterLevel(w, config.Level)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer gz.Close()

			// Wrap response writer
			grw := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
				config:         config,
			}

			// Set Content-Encoding header
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length") // Remove Content-Length as it will change

			next.ServeHTTP(grw, r)
		})
	}
}

// CompressWithDefault returns a Compress middleware with default configuration.
func CompressWithDefault() func(http.Handler) http.Handler {
	return Compress(DefaultCompressConfig())
}

// gzipResponseWriter wraps http.ResponseWriter to compress the response.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
	config CompressConfig
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// Check if we should compress based on content type
	contentType := w.Header().Get("Content-Type")
	shouldCompress := false

	for _, ct := range w.config.ContentTypes {
		if strings.HasPrefix(contentType, ct) {
			shouldCompress = true
			break
		}
	}

	// If content type doesn't match or size is too small, write directly
	if !shouldCompress || len(b) < w.config.MinSize {
		return w.ResponseWriter.Write(b)
	}

	// Write compressed
	return w.Writer.Write(b)
}
