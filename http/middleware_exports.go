package http

import (
	"net/http"
	"time"

	"github.com/donnigundala/dgcore/http/middleware"
	"github.com/donnigundala/dgcore/logging"
)

// Middleware exports - make middleware accessible from http package

// Recovery middleware
func Recovery(logger *logging.Logger) func(http.Handler) http.Handler {
	return middleware.Recovery(logger)
}

func RecoveryWithDefault() func(http.Handler) http.Handler {
	return middleware.RecoveryWithDefault()
}

// CORS middleware
func CORS(config middleware.CORSConfig) func(http.Handler) http.Handler {
	return middleware.CORS(config)
}

func CORSWithDefault() func(http.Handler) http.Handler {
	return middleware.CORSWithDefault()
}

func DefaultCORSConfig() middleware.CORSConfig {
	return middleware.DefaultCORSConfig()
}

// Security Headers middleware
func SecurityHeaders(config middleware.SecurityConfig) func(http.Handler) http.Handler {
	return middleware.SecurityHeaders(config)
}

func SecurityHeadersWithDefault() func(http.Handler) http.Handler {
	return middleware.SecurityHeadersWithDefault()
}

func DefaultSecurityConfig() middleware.SecurityConfig {
	return middleware.DefaultSecurityConfig()
}

// Rate Limit middleware
func RateLimit(config middleware.RateLimitConfig) func(http.Handler) http.Handler {
	return middleware.RateLimit(config)
}

func RateLimitWithDefault() func(http.Handler) http.Handler {
	return middleware.RateLimitWithDefault()
}

func DefaultRateLimitConfig() middleware.RateLimitConfig {
	return middleware.DefaultRateLimitConfig()
}

// Timeout middleware
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(timeout)
}

// Body Size Limit middleware
func BodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return middleware.BodySizeLimit(maxBytes)
}

func BodySizeLimitWithError(maxBytes int64) func(http.Handler) http.Handler {
	return middleware.BodySizeLimitWithError(maxBytes)
}

// Compress middleware
func Compress(config middleware.CompressConfig) func(http.Handler) http.Handler {
	return middleware.Compress(config)
}

func CompressWithDefault() func(http.Handler) http.Handler {
	return middleware.CompressWithDefault()
}

func DefaultCompressConfig() middleware.CompressConfig {
	return middleware.DefaultCompressConfig()
}
