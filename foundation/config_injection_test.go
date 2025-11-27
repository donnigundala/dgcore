package foundation

import (
	"testing"

	"github.com/donnigundala/dg-core/config"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type TestConfigWithValidation struct {
	Name string `mapstructure:"name" validate:"required"`
	Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
}

type TestProviderWithConfig struct {
	Config TestConfig `config:"test"`
}

type TestProviderWithValidation struct {
	Config TestConfigWithValidation `config:"test" validate:"required"`
}

type TestProviderMultipleConfigs struct {
	AppConfig TestConfig `config:"app"`
	DBConfig  TestConfig `config:"db"`
}

type TestProviderNoConfig struct {
	Name string
}

func TestInjectProviderConfig_Success(t *testing.T) {
	// Setup config
	config.Add("test", map[string]any{
		"name": "myapp",
		"port": 8080,
	})

	// Inject
	provider := &TestProviderWithConfig{}
	err := InjectProviderConfig(provider)

	assert.NoError(t, err)
	assert.Equal(t, "myapp", provider.Config.Name)
	assert.Equal(t, 8080, provider.Config.Port)
}

func TestInjectProviderConfig_WithValidation_Success(t *testing.T) {
	// Setup valid config
	config.Add("test", map[string]any{
		"name": "myapp",
		"port": 8080,
	})

	// Inject
	provider := &TestProviderWithValidation{}
	err := InjectProviderConfig(provider)

	// Validation is optional - if validator not available, it should still inject
	if err != nil && !assert.Contains(t, err.Error(), "validation is not available") {
		t.Fatalf("Unexpected error: %v", err)
	}

	// If no error, verify values were injected
	if err == nil {
		assert.Equal(t, "myapp", provider.Config.Name)
		assert.Equal(t, 8080, provider.Config.Port)
	}
}

func TestInjectProviderConfig_WithValidation_Failure(t *testing.T) {
	// Setup invalid config (missing required field)
	config.Add("test", map[string]any{
		"port": 8080,
		// name is missing
	})

	// Inject
	provider := &TestProviderWithValidation{}
	err := InjectProviderConfig(provider)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to inject config")
}

func TestInjectProviderConfig_MultipleFields(t *testing.T) {
	// Setup config for multiple fields
	config.Add("app", map[string]any{
		"name": "app1",
		"port": 3000,
	})
	config.Add("db", map[string]any{
		"name": "postgres",
		"port": 5432,
	})

	// Inject
	provider := &TestProviderMultipleConfigs{}
	err := InjectProviderConfig(provider)

	assert.NoError(t, err)
	assert.Equal(t, "app1", provider.AppConfig.Name)
	assert.Equal(t, 3000, provider.AppConfig.Port)
	assert.Equal(t, "postgres", provider.DBConfig.Name)
	assert.Equal(t, 5432, provider.DBConfig.Port)
}

func TestInjectProviderConfig_NoConfigTag(t *testing.T) {
	// Provider without config tag should not error
	provider := &TestProviderNoConfig{Name: "test"}
	err := InjectProviderConfig(provider)

	assert.NoError(t, err)
	assert.Equal(t, "test", provider.Name) // Unchanged
}

func TestInjectProviderConfig_NonStruct(t *testing.T) {
	// Non-struct should be skipped without error
	var notAStruct string = "test"
	err := InjectProviderConfig(&notAStruct)

	assert.NoError(t, err)
}

func TestInjectProviderConfig_MissingConfig(t *testing.T) {
	// Config key doesn't exist in config
	provider := &TestProviderWithConfig{}
	// Don't set any config

	err := InjectProviderConfig(provider)

	// Should not error - config.Inject handles missing keys gracefully
	// (returns zero values)
	assert.NoError(t, err)
}
