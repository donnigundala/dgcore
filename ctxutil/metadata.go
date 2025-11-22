package ctxutil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP extracts the client IP address from the request.
// It checks X-Forwarded-For, X-Real-IP, and RemoteAddr in that order.
func ClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// UserAgent returns the User-Agent header from the request.
func UserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// RequestPath returns the request path.
func RequestPath(r *http.Request) string {
	return r.URL.Path
}

// RequestMethod returns the HTTP method.
func RequestMethod(r *http.Request) string {
	return r.Method
}

// RequestHost returns the request host.
func RequestHost(r *http.Request) string {
	return r.Host
}

// RequestScheme returns the request scheme (http or https).
func RequestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	// Check X-Forwarded-Proto header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}

// RequestURL returns the full request URL.
func RequestURL(r *http.Request) string {
	scheme := RequestScheme(r)
	return scheme + "://" + r.Host + r.RequestURI
}
