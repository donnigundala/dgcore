package testing

import (
	"testing"

	"github.com/donnigundala/dg-core/contracts/foundation"
	"github.com/stretchr/testify/assert"
)

// MockProvider for testing utilities
type MockProvider struct {
	RegisterCalled bool
	BootCalled     bool
}

func (p *MockProvider) Register(app foundation.Application) error {
	p.RegisterCalled = true
	app.Singleton("mock_service", func() interface{} {
		return "mock_value"
	})
	return nil
}

func (p *MockProvider) Boot(app foundation.Application) error {
	p.BootCalled = true
	return nil
}

func TestProviderUtilities(t *testing.T) {
	// 1. Test NewTestApp
	app := NewTestApp()
	assert.NotNil(t, app)

	// 2. Test RegisterAndBoot
	provider := &MockProvider{}
	RegisterAndBoot(t, app, provider)

	assert.True(t, provider.RegisterCalled)
	assert.True(t, provider.BootCalled)

	// 3. Test AssertBound
	AssertBound(t, app, "mock_service")

	// 4. Test AssertResolved
	instance := AssertResolved(t, app, "mock_service")
	assert.Equal(t, "mock_value", instance)
}
