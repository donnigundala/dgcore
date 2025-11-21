package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/donnigundala/dgcore/contracts/foundation"
	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
	dghttp "github.com/donnigundala/dgcore/http"
)

// MockApplication is a mock implementation of foundation.Application
type MockApplication struct {
	foundation.Application // Embed to satisfy interface, but we only implement what we need
	booted                 bool
}

func (m *MockApplication) Boot() {
	m.booted = true
}

func (m *MockApplication) IsBooted() bool {
	return m.booted
}

// MockRouter is a mock implementation of contractHTTP.Router
type MockRouter struct {
	contractHTTP.Router // Embed to satisfy interface
	ServeHTTPFunc       func(w http.ResponseWriter, r *http.Request)
}

func (m *MockRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.ServeHTTPFunc != nil {
		m.ServeHTTPFunc(w, r)
	}
}

func TestKernel_Bootstrap(t *testing.T) {
	app := &MockApplication{}
	router := &MockRouter{}
	kernel := dghttp.NewKernel(app, router)

	if app.IsBooted() {
		t.Error("App should not be booted yet")
	}

	kernel.Bootstrap()

	if !app.IsBooted() {
		t.Error("App should be booted after Kernel.Bootstrap()")
	}
}

func TestKernel_Handle(t *testing.T) {
	app := &MockApplication{}
	router := &MockRouter{
		ServeHTTPFunc: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Router Handled"))
		},
	}
	kernel := dghttp.NewKernel(app, router)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	kernel.Handle(w, req)

	if w.Body.String() != "Router Handled" {
		t.Errorf("Expected 'Router Handled', got '%s'", w.Body.String())
	}

	if !app.IsBooted() {
		t.Error("App should be booted after Kernel.Handle()")
	}
}
