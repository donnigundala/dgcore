package foundation_test

import (
	"testing"

	"github.com/donnigundala/dgcore/contracts/container"
	foundationImpl "github.com/donnigundala/dgcore/foundation"
)

// MockService is a simple service for testing.
type MockService struct {
	Value string
}

// MockServiceProvider is a simple service provider for testing.
type MockServiceProvider struct {
	App container.Container
}

func (p *MockServiceProvider) Register() {
	p.App.Bind("mock_service", func() interface{} {
		return &MockService{}
	})
}

func (p *MockServiceProvider) Boot() {
	// Boot logic
}

func TestApplicationLifecycle(t *testing.T) {
	app := foundationImpl.New("/tmp")

	// Verify BasePath
	if app.BasePath() != "/tmp" {
		t.Errorf("Expected BasePath to be /tmp, got %s", app.BasePath())
	}

	// Verify Container is embedded
	if app.Container == nil {
		t.Error("Container should not be nil")
	}

	// Test Provider Registration
	provider := &MockServiceProvider{App: app}
	app.Register(provider)

	// Verify service is bound
	instance, err := app.Make("mock_service")
	if err != nil {
		t.Errorf("Failed to resolve mock_service: %v", err)
	}

	if _, ok := instance.(*MockService); !ok {
		t.Error("Resolved instance is not of type *MockService")
	}
}
