package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/donnigundala/dg-core/http/middleware"
)

// TestCORS_AllowedOrigins tests CORS with allowed origins
func TestCORS_AllowedOrigins(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	config.AllowedOrigins = []string{"https://example.com"}

	corsMiddleware := middleware.CORS(config)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin header to be 'https://example.com', got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestCORS_DisallowedOrigin tests CORS with disallowed origin
func TestCORS_DisallowedOrigin(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	config.AllowedOrigins = []string{"https://example.com"}

	corsMiddleware := middleware.CORS(config)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should not set CORS headers for disallowed origin
	if w.Header().Get("Access-Control-Allow-Origin") == "https://evil.com" {
		t.Error("Should not allow disallowed origin")
	}
}

// TestCORS_PreflightRequest tests CORS preflight (OPTIONS) request
func TestCORS_PreflightRequest(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	config.AllowedOrigins = []string{"*"}
	config.AllowedMethods = []string{"GET", "POST", "PUT"}

	corsMiddleware := middleware.CORS(config)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 204 No Content for preflight
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("Expected status 204 or 200 for preflight, got %d", w.Code)
	}
}

// TestCORS_Credentials tests CORS with credentials
func TestCORS_Credentials(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	config.AllowedOrigins = []string{"https://example.com"}
	config.AllowCredentials = true

	corsMiddleware := middleware.CORS(config)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected Access-Control-Allow-Credentials header to be 'true'")
	}
}

// TestCORS_AllowedHeaders tests CORS with allowed headers
func TestCORS_AllowedHeaders(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	config.AllowedOrigins = []string{"*"}
	config.AllowedHeaders = []string{"Content-Type", "Authorization"}

	corsMiddleware := middleware.CORS(config)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowedHeaders == "" {
		t.Error("Expected Access-Control-Allow-Headers header to be set")
	}
}

// TestCORSWithDefault tests the default CORS middleware
func TestCORSWithDefault(t *testing.T) {
	corsMiddleware := middleware.CORSWithDefault()

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Default config should allow all origins
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}
}

// TestCORS_Wildcard tests CORS with wildcard origin
func TestCORS_Wildcard(t *testing.T) {
	config := middleware.DefaultCORSConfig()
	config.AllowedOrigins = []string{"*"}

	corsMiddleware := middleware.CORS(config)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set for wildcard")
	}
}
