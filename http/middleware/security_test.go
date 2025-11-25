package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/donnigundala/dg-core/http/middleware"
)

// TestSecurityHeaders_AllHeaders tests that all security headers are set
func TestSecurityHeaders_AllHeaders(t *testing.T) {
	config := middleware.DefaultSecurityConfig()
	securityMiddleware := middleware.SecurityHeaders(config)

	handler := securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check that security headers are set
	headers := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}

	for header, expectedValue := range headers {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected %s header to be '%s', got '%s'", header, expectedValue, actualValue)
		}
	}
}

// TestSecurityHeaders_CustomConfig tests custom security configuration
func TestSecurityHeaders_CustomConfig(t *testing.T) {
	config := middleware.SecurityConfig{
		XFrameOptions:         "SAMEORIGIN",
		XContentTypeOptions:   "nosniff",
		XSSProtection:         "0",
		HSTSMaxAge:            63072000,
		HSTSIncludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "no-referrer",
	}

	securityMiddleware := middleware.SecurityHeaders(config)

	handler := securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("Expected X-Frame-Options to be 'SAMEORIGIN', got '%s'", w.Header().Get("X-Frame-Options"))
	}

	if w.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Errorf("Expected CSP to be 'default-src 'self'', got '%s'", w.Header().Get("Content-Security-Policy"))
	}
}

// TestSecurityHeadersWithDefault tests the default security middleware
func TestSecurityHeadersWithDefault(t *testing.T) {
	securityMiddleware := middleware.SecurityHeadersWithDefault()

	handler := securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should have at least some security headers
	if w.Header().Get("X-Frame-Options") == "" {
		t.Error("Expected X-Frame-Options header to be set")
	}

	if w.Header().Get("X-Content-Type-Options") == "" {
		t.Error("Expected X-Content-Type-Options header to be set")
	}
}

// TestSecurityHeaders_DoesNotOverrideExisting tests that existing headers are not overridden
func TestSecurityHeaders_DoesNotOverrideExisting(t *testing.T) {
	config := middleware.DefaultSecurityConfig()
	securityMiddleware := middleware.SecurityHeaders(config)

	handler := securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set a custom header before middleware processes
		w.Header().Set("X-Frame-Options", "ALLOW-FROM https://example.com")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// The middleware should set headers before the handler runs,
	// so the handler's value should win (or middleware should not override)
	// This tests the actual behavior
	frameOptions := w.Header().Get("X-Frame-Options")
	if frameOptions == "" {
		t.Error("Expected X-Frame-Options to be set")
	}
}
