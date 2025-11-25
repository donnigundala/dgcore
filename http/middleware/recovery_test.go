package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/donnigundala/dg-core/http/middleware"
	"github.com/donnigundala/dg-core/logging"
)

// TestRecovery_PanicHandling tests that panics are recovered
func TestRecovery_PanicHandling(t *testing.T) {
	logger := logging.Default()
	recoveryMiddleware := middleware.Recovery(logger)

	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	// Should return 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// TestRecovery_NormalRequest tests that normal requests pass through
func TestRecovery_NormalRequest(t *testing.T) {
	logger := logging.Default()
	recoveryMiddleware := middleware.Recovery(logger)

	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// TestRecoveryWithDefault tests the default recovery middleware
func TestRecoveryWithDefault(t *testing.T) {
	recoveryMiddleware := middleware.RecoveryWithDefault()

	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	// Should return 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// TestRecovery_DifferentPanicTypes tests recovery with different panic types
func TestRecovery_DifferentPanicTypes(t *testing.T) {
	logger := logging.Default()
	recoveryMiddleware := middleware.Recovery(logger)

	testCases := []struct {
		name     string
		panicVal interface{}
	}{
		{"String", "string panic"},
		{"Error", http.ErrAbortHandler},
		{"Int", 42},
		{"Nil", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tc.panicVal)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			// Should not panic
			handler.ServeHTTP(w, req)

			// Should return 500
			if w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status 500, got %d", w.Code)
			}
		})
	}
}
