package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"github.com/donnigundala/dg-core/errors"
)

// RateLimitConfig defines the configuration for rate limiting middleware.
type RateLimitConfig struct {
	// RequestsPerSecond is the number of requests allowed per second.
	// Default: 10
	RequestsPerSecond float64

	// BurstSize is the maximum burst size.
	// Default: 20
	BurstSize int

	// KeyFunc is a function to extract the key for rate limiting (e.g., IP address).
	// Default: uses client IP
	KeyFunc func(*http.Request) string
}

// DefaultRateLimitConfig returns the default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
		KeyFunc:           getClientIP,
	}
}

// RateLimit returns a middleware that limits requests per client.
func RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	// Apply defaults
	if config.RequestsPerSecond == 0 {
		config.RequestsPerSecond = 10
	}
	if config.BurstSize == 0 {
		config.BurstSize = 20
	}
	if config.KeyFunc == nil {
		config.KeyFunc = getClientIP
	}

	// Create a map to store limiters per client
	limiters := &sync.Map{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client key
			key := config.KeyFunc(r)

			// Get or create limiter for this client
			limiterInterface, _ := limiters.LoadOrStore(key, rate.NewLimiter(
				rate.Limit(config.RequestsPerSecond),
				config.BurstSize,
			))
			limiter := limiterInterface.(*rate.Limiter)

			// Check if request is allowed
			if !limiter.Allow() {
				err := errors.ErrTooManyRequests.
					WithField("client", key)
				errors.WriteHTTPError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitWithDefault returns a RateLimit middleware with default configuration.
func RateLimitWithDefault() func(http.Handler) http.Handler {
	return RateLimit(DefaultRateLimitConfig())
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if ip := parseForwardedFor(xff); ip != "" {
			return ip
		}
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// parseForwardedFor parses the X-Forwarded-For header and returns the first IP.
func parseForwardedFor(xff string) string {
	// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
	// We want the first one (the client)
	for idx := 0; idx < len(xff); idx++ {
		if xff[idx] == ',' {
			return xff[:idx]
		}
	}
	return xff
}
