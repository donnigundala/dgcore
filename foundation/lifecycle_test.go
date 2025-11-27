package foundation

import (
	"context"
	"errors"
	"testing"

	"github.com/donnigundala/dg-core/contracts/foundation"
	"github.com/stretchr/testify/assert"
)

// LifecycleProvider implements all lifecycle hooks for testing
type LifecycleProvider struct {
	MockPlugin // Embed MockPlugin for basic Provider implementation

	beforeRegisterCalled bool
	afterBootCalled      bool
	shutdownCalled       bool

	shouldFailBeforeRegister bool
	shouldFailAfterBoot      bool
	shouldFailShutdown       bool
}

func (p *LifecycleProvider) BeforeRegister(app foundation.Application) error {
	if p.shouldFailBeforeRegister {
		return errors.New("before register failed")
	}
	p.beforeRegisterCalled = true
	return nil
}

func (p *LifecycleProvider) AfterBoot(app foundation.Application) error {
	if p.shouldFailAfterBoot {
		return errors.New("after boot failed")
	}
	p.afterBootCalled = true
	return nil
}

func (p *LifecycleProvider) Shutdown(app foundation.Application) error {
	if p.shouldFailShutdown {
		return errors.New("shutdown failed")
	}
	p.shutdownCalled = true
	return nil
}

func TestLifecycle_BeforeRegister(t *testing.T) {
	app := New("/tmp/test")
	provider := &LifecycleProvider{MockPlugin: MockPlugin{name: "lifecycle"}}

	err := app.Register(provider)

	assert.NoError(t, err)
	assert.True(t, provider.beforeRegisterCalled)
	assert.True(t, provider.registered)
}

func TestLifecycle_BeforeRegister_Failure(t *testing.T) {
	app := New("/tmp/test")
	provider := &LifecycleProvider{
		MockPlugin:               MockPlugin{name: "lifecycle"},
		shouldFailBeforeRegister: true,
	}

	err := app.Register(provider)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BeforeRegister hook failed")
	assert.False(t, provider.registered) // Should not proceed to Register
}

func TestLifecycle_AfterBoot(t *testing.T) {
	app := New("/tmp/test")
	provider := &LifecycleProvider{MockPlugin: MockPlugin{name: "lifecycle"}}

	app.Register(provider)
	err := app.Boot()

	assert.NoError(t, err)
	assert.True(t, provider.booted)
	assert.True(t, provider.afterBootCalled)
}

func TestLifecycle_AfterBoot_Failure(t *testing.T) {
	app := New("/tmp/test")
	provider := &LifecycleProvider{
		MockPlugin:          MockPlugin{name: "lifecycle"},
		shouldFailAfterBoot: true,
	}

	app.Register(provider)
	err := app.Boot()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AfterBoot hook failed")
}

func TestLifecycle_Shutdown(t *testing.T) {
	app := New("/tmp/test")
	provider := &LifecycleProvider{MockPlugin: MockPlugin{name: "lifecycle"}}

	app.Register(provider)

	// Trigger shutdown
	err := app.Shutdown(context.Background())

	assert.NoError(t, err)
	assert.True(t, provider.shutdownCalled)
}

func TestLifecycle_LateRegistration(t *testing.T) {
	// Test registering a provider AFTER app is already booted
	app := New("/tmp/test")
	app.Boot()

	provider := &LifecycleProvider{MockPlugin: MockPlugin{name: "late-provider"}}

	err := app.Register(provider)

	assert.NoError(t, err)
	assert.True(t, provider.beforeRegisterCalled)
	assert.True(t, provider.registered)
	assert.True(t, provider.booted)
	assert.True(t, provider.afterBootCalled)
}
