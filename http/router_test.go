package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contractHTTP "github.com/donnigundala/dg-core/contracts/http"
	dghttp "github.com/donnigundala/dg-core/http"
)

func TestRouter_BasicRoutes(t *testing.T) {
	router := dghttp.NewRouter()

	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	router.Post("/submit", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Submitted"))
	})

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", w.Body.String())
	}

	// Test POST
	req = httptest.NewRequest(http.MethodPost, "/submit", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "Submitted" {
		t.Errorf("Expected 'Submitted', got '%s'", w.Body.String())
	}

	// Test NotFound
	req = httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestRouter_Parameters(t *testing.T) {
	router := dghttp.NewRouter()

	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, we would extract params from context
		w.Write([]byte("User ID"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "User ID" {
		t.Errorf("Expected 'User ID', got '%s'", w.Body.String())
	}
}

func TestRouter_Groups(t *testing.T) {
	router := dghttp.NewRouter()

	router.Group(contractHTTP.GroupAttributes{Prefix: "/api"}, func(r contractHTTP.Router) {
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API Users"))
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "API Users" {
		t.Errorf("Expected 'API Users', got '%s'", w.Body.String())
	}
}

func TestRouter_Middleware(t *testing.T) {
	router := dghttp.NewRouter()

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "Passed")
			next.ServeHTTP(w, r)
		})
	}

	router.Use(middleware)

	router.Get("/middleware", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Middleware Test"))
	})

	req := httptest.NewRequest(http.MethodGet, "/middleware", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-Test") != "Passed" {
		t.Errorf("Expected X-Test header to be 'Passed', got '%s'", w.Header().Get("X-Test"))
	}
}
