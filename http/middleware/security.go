package middleware

import (
	"fmt"
	"net/http"
)

// SecurityConfig defines the configuration for security headers middleware.
type SecurityConfig struct {
	// ContentSecurityPolicy sets the Content-Security-Policy header.
	// Default: "default-src 'self'"
	ContentSecurityPolicy string

	// XFrameOptions sets the X-Frame-Options header.
	// Default: "DENY"
	XFrameOptions string

	// XContentTypeOptions sets the X-Content-Type-Options header.
	// Default: "nosniff"
	XContentTypeOptions string

	// XSSProtection sets the X-XSS-Protection header.
	// Default: "1; mode=block"
	XSSProtection string

	// HSTSMaxAge sets the max-age for Strict-Transport-Security header (in seconds).
	// Default: 31536000 (1 year)
	// Set to 0 to disable HSTS
	HSTSMaxAge int

	// HSTSIncludeSubdomains includes subdomains in HSTS.
	// Default: false
	HSTSIncludeSubdomains bool

	// ReferrerPolicy sets the Referrer-Policy header.
	// Default: "strict-origin-when-cross-origin"
	ReferrerPolicy string
}

// DefaultSecurityConfig returns the default security configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ContentSecurityPolicy: "default-src 'self'",
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
		XSSProtection:         "1; mode=block",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: false,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// SecurityHeaders returns a middleware that adds security headers to responses.
func SecurityHeaders(config SecurityConfig) func(http.Handler) http.Handler {
	// Apply defaults
	if config.ContentSecurityPolicy == "" {
		config.ContentSecurityPolicy = "default-src 'self'"
	}
	if config.XFrameOptions == "" {
		config.XFrameOptions = "DENY"
	}
	if config.XContentTypeOptions == "" {
		config.XContentTypeOptions = "nosniff"
	}
	if config.XSSProtection == "" {
		config.XSSProtection = "1; mode=block"
	}
	if config.ReferrerPolicy == "" {
		config.ReferrerPolicy = "strict-origin-when-cross-origin"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// X-Frame-Options
			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			}

			// X-Content-Type-Options
			if config.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.XContentTypeOptions)
			}

			// X-XSS-Protection
			if config.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", config.XSSProtection)
			}

			// Strict-Transport-Security (HSTS)
			if config.HSTSMaxAge > 0 {
				hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
				if config.HSTSIncludeSubdomains {
					hstsValue += "; includeSubDomains"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Referrer-Policy
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersWithDefault returns a SecurityHeaders middleware with default configuration.
func SecurityHeadersWithDefault() func(http.Handler) http.Handler {
	return SecurityHeaders(DefaultSecurityConfig())
}
