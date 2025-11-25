package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/donnigundala/dg-core/http/middleware"
	"github.com/donnigundala/dg-core/logging"
)

// TestLogger_BasicRequest tests basic request logging
func TestLogger_BasicRequest(t *testing.T) {
	logger := logging.Default()

	config := middleware.DefaultLoggerConfig()
	config.Logger = logger

	loggerMiddleware := middleware.Logger(config)

	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected 'OK', got '%s'", w.Body.String())
	}
}

// TestLogger_WithQueryParams tests logging with query parameters
func TestLogger_WithQueryParams(t *testing.T) {
	logger := logging.Default()

	config := middleware.DefaultLoggerConfig()
	config.Logger = logger

	loggerMiddleware := middleware.Logger(config)

	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar&baz=qux", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestLogger_SkipPaths tests that certain paths are skipped
func TestLogger_SkipPaths(t *testing.T) {
	logger := logging.Default()

	config := middleware.DefaultLoggerConfig()
	config.Logger = logger
	config.SkipPaths = []string{"/health", "/metrics"}

	loggerMiddleware := middleware.Logger(config)

	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test skipped path
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test non-skipped path
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestLogger_ClientIP tests client IP detection
func TestLogger_ClientIP_XForwardedFor(t *testing.T) {
	logger := logging.Default()

	config := middleware.DefaultLoggerConfig()
	config.Logger = logger
	config.LogClientIP = true

	loggerMiddleware := middleware.Logger(config)

	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestLogger_StatusCodes tests logging different status codes
func TestLogger_StatusCodes(t *testing.T) {
	logger := logging.Default()

	config := middleware.DefaultLoggerConfig()
	config.Logger = logger

	loggerMiddleware := middleware.Logger(config)

	testCases := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"InternalError", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, w.Code)
			}
		})
	}
}

// TestLoggerWithDefault tests the default logger middleware
func TestLoggerWithDefault(t *testing.T) {
	loggerMiddleware := middleware.LoggerWithDefault()

	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestLogger_ConfigOptions tests various configuration options
func TestLogger_ConfigOptions(t *testing.T) {
	logger := logging.Default()

	config := middleware.LoggerConfig{
		Logger:        logger,
		SkipPaths:     []string{"/skip"},
		LogLatency:    false,
		LogClientIP:   false,
		LogMethod:     true,
		LogPath:       true,
		LogStatusCode: true,
	}

	loggerMiddleware := middleware.Logger(config)

	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"name":"test"}`))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
