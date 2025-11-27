package foundation

import (
	"testing"

	"github.com/donnigundala/dg-core/contracts/foundation"
	"github.com/stretchr/testify/assert"
)

// MockPlugin is a mock plugin for testing
type MockPlugin struct {
	name         string
	version      string
	dependencies []string
	registered   bool
	booted       bool
}

func (m *MockPlugin) Name() string {
	return m.name
}

func (m *MockPlugin) Version() string {
	return m.version
}

func (m *MockPlugin) Dependencies() []string {
	return m.dependencies
}

func (m *MockPlugin) Register(app foundation.Application) error {
	m.registered = true
	return nil
}

func (m *MockPlugin) Boot(app foundation.Application) error {
	m.booted = true
	return nil
}

func TestRegisterPlugin_Success(t *testing.T) {
	app := New("/tmp/test")
	plugin := &MockPlugin{name: "test-plugin", version: "1.0.0"}

	err := app.RegisterPlugin(plugin)

	assert.NoError(t, err)
	assert.True(t, app.HasPlugin("test-plugin"))
	assert.True(t, plugin.registered)
}

func TestRegisterPlugin_Duplicate(t *testing.T) {
	app := New("/tmp/test")
	plugin := &MockPlugin{name: "test-plugin", version: "1.0.0"}

	// First registration should succeed
	err := app.RegisterPlugin(plugin)
	assert.NoError(t, err)

	// Second registration should fail
	err = app.RegisterPlugin(plugin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
	assert.Contains(t, err.Error(), "test-plugin")
}

func TestRegisterPlugin_DifferentPlugins(t *testing.T) {
	app := New("/tmp/test")

	plugin1 := &MockPlugin{name: "plugin1", version: "1.0.0"}
	plugin2 := &MockPlugin{name: "plugin2", version: "1.0.0"}

	// Both should succeed
	err := app.RegisterPlugin(plugin1)
	assert.NoError(t, err)

	err = app.RegisterPlugin(plugin2)
	assert.NoError(t, err)

	assert.True(t, app.HasPlugin("plugin1"))
	assert.True(t, app.HasPlugin("plugin2"))
}

func TestRegisterPlugin_SameNameDifferentInstance(t *testing.T) {
	app := New("/tmp/test")

	plugin1 := &MockPlugin{name: "same-name", version: "1.0.0"}
	plugin2 := &MockPlugin{name: "same-name", version: "2.0.0"}

	// First registration should succeed
	err := app.RegisterPlugin(plugin1)
	assert.NoError(t, err)

	// Second registration with same name should fail
	err = app.RegisterPlugin(plugin2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}
