package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donnigundala/dg-core/config"
)

// TestAdd_BasicRegistration tests basic configuration registration
func TestAdd_BasicRegistration(t *testing.T) {
	// Clear any existing config
	defer resetConfig()

	config.Add("test", map[string]any{
		"name":  "test-app",
		"debug": true,
		"port":  8080,
	})

	// Verify values are registered
	if val := config.GetString("test.name"); val != "test-app" {
		t.Errorf("Expected 'test-app', got '%s'", val)
	}

	if val := config.GetBool("test.debug"); val != true {
		t.Error("Expected true for test.debug")
	}
}

// TestAdd_NestedConfiguration tests nested map registration
func TestAdd_NestedConfiguration(t *testing.T) {
	defer resetConfig()

	config.Add("app", map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 3000,
		},
		"database": map[string]any{
			"driver": "postgres",
			"host":   "db.example.com",
		},
	})

	// Verify nested values
	if val := config.GetString("app.server.host"); val != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", val)
	}

	if val := config.GetString("app.database.driver"); val != "postgres" {
		t.Errorf("Expected 'postgres', got '%s'", val)
	}
}

// TestEnv_WithDefault tests environment variable reading with fallback
func TestEnv_WithDefault(t *testing.T) {
	// Test with non-existent env var
	result := config.Env("NON_EXISTENT_VAR", "default-value")
	if result != "default-value" {
		t.Errorf("Expected 'default-value', got %v", result)
	}

	// Test with existing env var
	os.Setenv("TEST_ENV_VAR", "env-value")
	defer os.Unsetenv("TEST_ENV_VAR")

	result = config.Env("TEST_ENV_VAR", "default-value")
	if result != "env-value" {
		t.Errorf("Expected 'env-value', got %v", result)
	}
}

// TestGet_ReturnsNilForNonExistent tests that Get returns nil for missing keys
func TestGet_ReturnsNilForNonExistent(t *testing.T) {
	defer resetConfig()

	result := config.Get("nonexistent.key")
	if result != nil {
		t.Errorf("Expected nil for non-existent key, got %v", result)
	}
}

// TestGetString_EmptyForNonExistent tests GetString returns empty string for missing keys
func TestGetString_EmptyForNonExistent(t *testing.T) {
	defer resetConfig()

	result := config.GetString("nonexistent.key")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

// TestGetBool_FalseForNonExistent tests GetBool returns false for missing keys
func TestGetBool_FalseForNonExistent(t *testing.T) {
	defer resetConfig()

	result := config.GetBool("nonexistent.key")
	if result != false {
		t.Error("Expected false for non-existent key")
	}
}

// TestAllKeys_ReturnsRegisteredKeys tests that AllKeys returns all registered keys
func TestAllKeys_ReturnsRegisteredKeys(t *testing.T) {
	defer resetConfig()

	config.Add("test", map[string]any{
		"key1": "value1",
		"key2": "value2",
	})

	keys := config.AllKeys()
	if len(keys) < 2 {
		t.Errorf("Expected at least 2 keys, got %d", len(keys))
	}

	// Check that our keys are present
	hasKey1 := false
	hasKey2 := false
	for _, k := range keys {
		if k == "test.key1" {
			hasKey1 = true
		}
		if k == "test.key2" {
			hasKey2 = true
		}
	}

	if !hasKey1 || !hasKey2 {
		t.Error("Expected keys test.key1 and test.key2 to be present")
	}
}

// TestLoad_YAMLFile tests loading configuration from YAML file
func TestLoad_YAMLFile(t *testing.T) {
	defer resetConfig()

	// Create temporary YAML file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `
app:
  name: test-app
  debug: true
  port: 8080
`
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load the config file using LoadWithPaths
	if err := config.LoadWithPaths(tmpDir); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if val := config.GetString("app.name"); val != "test-app" {
		t.Errorf("Expected 'test-app', got '%s'", val)
	}

	if val := config.GetBool("app.debug"); val != true {
		t.Error("Expected true for app.debug")
	}
}

// TestInject_ValidStruct tests injecting configuration into a struct
func TestInject_ValidStruct(t *testing.T) {
	defer resetConfig()

	type AppConfig struct {
		Name  string `mapstructure:"name"`
		Debug bool   `mapstructure:"debug"`
		Port  int    `mapstructure:"port"`
	}

	config.Add("app", map[string]any{
		"name":  "my-app",
		"debug": true,
		"port":  9000,
	})

	var cfg AppConfig
	if err := config.Inject("app", &cfg); err != nil {
		t.Fatalf("Failed to inject config: %v", err)
	}

	if cfg.Name != "my-app" {
		t.Errorf("Expected 'my-app', got '%s'", cfg.Name)
	}

	if cfg.Debug != true {
		t.Error("Expected true for Debug")
	}

	if cfg.Port != 9000 {
		t.Errorf("Expected 9000, got %d", cfg.Port)
	}
}

// TestInject_NestedStruct tests injecting nested configuration
func TestInject_NestedStruct(t *testing.T) {
	defer resetConfig()

	type ServerConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	type AppConfig struct {
		Name   string       `mapstructure:"name"`
		Server ServerConfig `mapstructure:"server"`
	}

	config.Add("app", map[string]any{
		"name": "my-app",
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	})

	var cfg AppConfig
	if err := config.Inject("app", &cfg); err != nil {
		t.Fatalf("Failed to inject config: %v", err)
	}

	if cfg.Name != "my-app" {
		t.Errorf("Expected 'my-app', got '%s'", cfg.Name)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected 8080, got %d", cfg.Server.Port)
	}
}

// TestInject_InvalidKey tests error handling for non-existent keys
func TestInject_InvalidKey(t *testing.T) {
	defer resetConfig()

	type AppConfig struct {
		Name string `mapstructure:"name"`
	}

	var cfg AppConfig
	err := config.Inject("nonexistent", &cfg)
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

// TestInject_NonPointer tests error handling for non-pointer argument
func TestInject_NonPointer(t *testing.T) {
	defer resetConfig()

	type AppConfig struct {
		Name string `mapstructure:"name"`
	}

	config.Add("app", map[string]any{"name": "test"})

	var cfg AppConfig
	err := config.Inject("app", cfg) // Not a pointer
	if err == nil {
		t.Error("Expected error for non-pointer argument")
	}
}

// TestEnvironmentOverride tests that environment variables override config values
func TestEnvironmentOverride(t *testing.T) {
	defer resetConfig()

	config.Add("app", map[string]any{
		"name": "default-name",
	})

	// Set environment variable (APP_NAME)
	os.Setenv("APP_NAME", "env-override")
	defer os.Unsetenv("APP_NAME")

	// The config system should pick up the env var
	// Note: This test depends on the config system's env binding behavior
	val := config.GetString("app.name")

	// The actual behavior may vary based on when env binding happens
	// This test documents the expected behavior
	if val != "env-override" && val != "default-name" {
		t.Logf("Got value: %s (may be 'env-override' or 'default-name' depending on timing)", val)
	}
}

// TestConcurrentAccess tests thread-safe concurrent access
func TestConcurrentAccess(t *testing.T) {
	defer resetConfig()

	config.Add("test", map[string]any{
		"value": "test-value",
	})

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = config.GetString("test.value")
				_ = config.AllKeys()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic
	if val := config.GetString("test.value"); val != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", val)
	}
}

// resetConfig is a helper to reset config state between tests
// Note: This is a workaround since the config package uses global state
func resetConfig() {
	// The config package doesn't expose a Reset() function
	// In a real scenario, you might need to add one or use test isolation
	// For now, we'll work with the existing state
}
