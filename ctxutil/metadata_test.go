package ctxutil

import (
	"net/http/httptest"
	"testing"
)

// TestClientIP_XForwardedFor tests extracting IP from X-Forwarded-For header
func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

	ip := ClientIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", ip)
	}
}

// TestClientIP_XForwardedFor_Single tests single IP in X-Forwarded-For
func TestClientIP_XForwardedFor_Single(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")

	ip := ClientIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", ip)
	}
}

// TestClientIP_XRealIP tests extracting IP from X-Real-IP header
func TestClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.2")

	ip := ClientIP(req)
	if ip != "203.0.113.2" {
		t.Errorf("expected 203.0.113.2, got %s", ip)
	}
}

// TestClientIP_RemoteAddr tests fallback to RemoteAddr
func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.3:12345"

	ip := ClientIP(req)
	if ip != "203.0.113.3" {
		t.Errorf("expected 203.0.113.3, got %s", ip)
	}
}

// TestClientIP_Priority tests header priority (X-Forwarded-For > X-Real-IP > RemoteAddr)
func TestClientIP_Priority(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "203.0.113.2")
	req.RemoteAddr = "203.0.113.3:12345"

	ip := ClientIP(req)
	// Should use X-Forwarded-For first
	if ip != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1 (X-Forwarded-For), got %s", ip)
	}
}

// TestUserAgent tests retrieving User-Agent header
func TestUserAgent(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	ua := UserAgent(req)
	if ua != "Mozilla/5.0" {
		t.Errorf("expected Mozilla/5.0, got %s", ua)
	}
}

// TestUserAgent_Empty tests empty User-Agent
func TestUserAgent_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	ua := UserAgent(req)
	if ua != "" {
		t.Errorf("expected empty string, got %s", ua)
	}
}

// TestRequestPath tests retrieving request path
func TestRequestPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/users/123", nil)

	path := RequestPath(req)
	if path != "/api/users/123" {
		t.Errorf("expected /api/users/123, got %s", path)
	}
}

// TestRequestMethod tests retrieving HTTP method
func TestRequestMethod(t *testing.T) {
	tests := []struct {
		method string
	}{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)

			method := RequestMethod(req)
			if method != tt.method {
				t.Errorf("expected %s, got %s", tt.method, method)
			}
		})
	}
}

// TestRequestHost tests retrieving request host
func TestRequestHost(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/path", nil)

	host := RequestHost(req)
	if host != "example.com" {
		t.Errorf("expected example.com, got %s", host)
	}
}

// TestRequestScheme_HTTP tests HTTP scheme
func TestRequestScheme_HTTP(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)

	scheme := RequestScheme(req)
	if scheme != "http" {
		t.Errorf("expected http, got %s", scheme)
	}
}

// TestRequestScheme_XForwardedProto tests X-Forwarded-Proto header
func TestRequestScheme_XForwardedProto(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	scheme := RequestScheme(req)
	if scheme != "https" {
		t.Errorf("expected https, got %s", scheme)
	}
}

// TestRequestScheme_TLS tests HTTPS with TLS
func TestRequestScheme_TLS(t *testing.T) {
	req := httptest.NewRequest("GET", "https://example.com", nil)
	// In unit tests, we use X-Forwarded-Proto to simulate HTTPS
	req.Header.Set("X-Forwarded-Proto", "https")

	scheme := RequestScheme(req)
	if scheme != "https" {
		t.Errorf("expected https, got %s", scheme)
	}
}

// TestRequestURL tests full URL construction
func TestRequestURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/users?page=1", nil)
	req.Host = "example.com"

	url := RequestURL(req)
	expected := "http://example.com/api/users?page=1"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

// TestRequestURL_HTTPS tests HTTPS URL construction
func TestRequestURL_HTTPS(t *testing.T) {
	req := httptest.NewRequest("GET", "/secure", nil)
	req.Host = "example.com"
	req.Header.Set("X-Forwarded-Proto", "https")

	url := RequestURL(req)
	expected := "https://example.com/secure"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}
