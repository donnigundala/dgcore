package testing

import (
	"testing"

	"github.com/donnigundala/dg-core/contracts/foundation"
	coreFoundation "github.com/donnigundala/dg-core/foundation"
	"github.com/stretchr/testify/assert"
)

// NewTestApp creates a new application instance for testing.
func NewTestApp() *coreFoundation.Application {
	return coreFoundation.New("/tmp/test")
}

// RegisterAndBoot registers and boots a provider, failing the test on error.
func RegisterAndBoot(t *testing.T, app *coreFoundation.Application, provider foundation.ServiceProvider) {
	err := app.Register(provider)
	assert.NoError(t, err, "Failed to register provider")

	err = app.Boot()
	assert.NoError(t, err, "Failed to boot application")
}

// AssertBound asserts that a key is bound in the container.
func AssertBound(t *testing.T, app *coreFoundation.Application, key string) {
	_, err := app.Make(key)
	assert.NoError(t, err, "Expected %s to be bound", key)
}

// AssertResolved asserts that a key can be resolved from the container.
func AssertResolved(t *testing.T, app *coreFoundation.Application, key string) interface{} {
	instance, err := app.Make(key)
	assert.NoError(t, err, "Failed to resolve %s", key)
	assert.NotNil(t, instance, "Resolved instance for %s is nil", key)
	return instance
}
