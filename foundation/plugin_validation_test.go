package foundation_test

import (
	"testing"

	"github.com/donnigundala/dg-core/contracts/foundation"
	core "github.com/donnigundala/dg-core/foundation"
	"github.com/stretchr/testify/assert"
)

// mockPlugin implements foundation.PluginProvider
type mockPlugin struct {
	name string
	deps []string
}

func (p *mockPlugin) Register(app foundation.Application) error { return nil }
func (p *mockPlugin) Boot(app foundation.Application) error     { return nil }
func (p *mockPlugin) Name() string                              { return p.name }
func (p *mockPlugin) Version() string                           { return "1.0.0" }
func (p *mockPlugin) Dependencies() []string                    { return p.deps }

func TestApplication_PluginDependencyValidation(t *testing.T) {
	t.Run("Boot fails when dependency is missing", func(t *testing.T) {
		app := core.New(".")

		// Register plugin A that depends on plugin B
		pluginA := &mockPlugin{
			name: "plugin-a",
			deps: []string{"plugin-b"},
		}

		err := app.RegisterPlugin(pluginA)
		assert.NoError(t, err)

		// Boot should fail because plugin-b is missing
		err = app.Boot()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin 'plugin-a' requires dependency 'plugin-b'")
	})

	t.Run("Boot succeeds when dependency is present", func(t *testing.T) {
		app := core.New(".")

		// Register plugin B (the dependency)
		pluginB := &mockPlugin{
			name: "plugin-b",
		}
		err := app.RegisterPlugin(pluginB)
		assert.NoError(t, err)

		// Register plugin A that depends on plugin B
		pluginA := &mockPlugin{
			name: "plugin-a",
			deps: []string{"plugin-b"},
		}
		err = app.RegisterPlugin(pluginA)
		assert.NoError(t, err)

		// Boot should succeed
		err = app.Boot()
		assert.NoError(t, err)
	})
}
