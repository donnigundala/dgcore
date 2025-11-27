package container_test

import (
	"sync"
	"testing"

	"github.com/donnigundala/dg-core/container"
)

// TestBind_Basic tests basic binding functionality
func TestBind_Basic(t *testing.T) {
	c := container.NewContainer()

	// Bind a simple value
	c.Bind("test", func() interface{} {
		return "test value"
	})

	// Resolve the binding
	result, err := c.Make("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "test value" {
		t.Errorf("Expected 'test value', got %v", result)
	}
}

// TestBind_WithFunction tests binding with different function signatures
func TestBind_WithFunction(t *testing.T) {
	c := container.NewContainer()

	// Bind with no-arg function
	c.Bind("simple", func() interface{} {
		return 42
	})

	result, err := c.Make("simple")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

// TestBind_MultipleInstances tests that Bind creates new instances each time
func TestBind_MultipleInstances(t *testing.T) {
	c := container.NewContainer()

	counter := 0
	c.Bind("counter", func() interface{} {
		counter++
		return counter
	})

	// First call
	result1, _ := c.Make("counter")
	// Second call
	result2, _ := c.Make("counter")

	// Should create new instances each time
	if result1 == result2 {
		t.Error("Expected different instances, got same value")
	}

	if counter != 2 {
		t.Errorf("Expected counter to be 2, got %d", counter)
	}
}

// TestSingleton_SharedInstance tests that Singleton creates only one instance
func TestSingleton_SharedInstance(t *testing.T) {
	c := container.NewContainer()

	counter := 0
	c.Singleton("singleton", func() interface{} {
		counter++
		return &struct{ Value int }{Value: counter}
	})

	// First call
	result1, err := c.Make("singleton")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second call
	result2, err := c.Make("singleton")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return the same instance
	if result1 != result2 {
		t.Error("Expected same instance for singleton")
	}

	// Counter should only be incremented once
	if counter != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter)
	}
}

// TestSingleton_ThreadSafe tests thread safety of singleton resolution
func TestSingleton_ThreadSafe(t *testing.T) {
	c := container.NewContainer()

	counter := 0
	var mu sync.Mutex

	c.Singleton("concurrent", func() interface{} {
		mu.Lock()
		defer mu.Unlock()
		counter++
		return &struct{ ID int }{ID: counter}
	})

	// Resolve concurrently
	const goroutines = 100
	var wg sync.WaitGroup
	results := make([]interface{}, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := c.Make("concurrent")
			if err != nil {
				t.Errorf("Error in goroutine %d: %v", index, err)
			}
			results[index] = result
		}(i)
	}

	wg.Wait()

	// All results should be the same instance
	firstResult := results[0]
	for i, result := range results {
		if result != firstResult {
			t.Errorf("Result %d is different from first result", i)
		}
	}

	// Counter should only be 1 (singleton created once)
	if counter != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter)
	}
}

// TestInstance_DirectBinding tests direct instance binding
func TestInstance_DirectBinding(t *testing.T) {
	c := container.NewContainer()

	type TestStruct struct {
		Name string
	}

	instance := &TestStruct{Name: "test"}
	c.Instance("direct", instance)

	result, err := c.Make("direct")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return the exact same instance
	if result != instance {
		t.Error("Expected same instance")
	}

	resultStruct, ok := result.(*TestStruct)
	if !ok {
		t.Fatal("Expected *TestStruct type")
	}

	if resultStruct.Name != "test" {
		t.Errorf("Expected name 'test', got %s", resultStruct.Name)
	}
}

// TestMake_NotFound tests error handling for non-existent bindings
func TestMake_NotFound(t *testing.T) {
	c := container.NewContainer()

	_, err := c.Make("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent binding")
	}

	expectedMsg := "binding not found for key: nonexistent"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestFlush_ClearsAll tests that Flush removes all bindings and instances
func TestFlush_ClearsAll(t *testing.T) {
	c := container.NewContainer()

	// Add some bindings
	c.Bind("bind1", func() interface{} { return "value1" })
	c.Singleton("singleton1", func() interface{} { return "value2" })
	c.Instance("instance1", "value3")

	// Verify they exist
	if _, err := c.Make("bind1"); err != nil {
		t.Fatal("bind1 should exist before flush")
	}

	// Flush
	c.Flush()

	// Verify they're gone
	if _, err := c.Make("bind1"); err == nil {
		t.Error("bind1 should not exist after flush")
	}
	if _, err := c.Make("singleton1"); err == nil {
		t.Error("singleton1 should not exist after flush")
	}
	if _, err := c.Make("instance1"); err == nil {
		t.Error("instance1 should not exist after flush")
	}
}

// TestBind_OverwriteExisting tests that rebinding overwrites previous binding
func TestBind_OverwriteExisting(t *testing.T) {
	c := container.NewContainer()

	// First binding
	c.Bind("key", func() interface{} { return "first" })

	// Second binding (should overwrite)
	c.Bind("key", func() interface{} { return "second" })

	result, err := c.Make("key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "second" {
		t.Errorf("Expected 'second', got %v", result)
	}
}

// TestSingleton_AfterInstance tests that instances take precedence
func TestSingleton_AfterInstance(t *testing.T) {
	c := container.NewContainer()

	// Set instance first
	c.Instance("key", "instance value")

	// Verify instance is returned
	result, err := c.Make("key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "instance value" {
		t.Errorf("Expected 'instance value', got %v", result)
	}

	// Set singleton (instances take precedence, so this won't overwrite)
	c.Singleton("key", func() interface{} { return "singleton value" })

	result2, err := c.Make("key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Instance should still be returned (instances have priority)
	if result2 != "instance value" {
		t.Errorf("Expected 'instance value' (instances take precedence), got %v", result2)
	}
}

// TestConcurrentBindAndMake tests concurrent binding and resolution
func TestConcurrentBindAndMake(t *testing.T) {
	c := container.NewContainer()

	var wg sync.WaitGroup
	const operations = 50

	// Concurrent bindings
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := "key"
			c.Bind(key, func() interface{} { return index })
		}(i)
	}

	// Concurrent resolutions
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Make("key")
		}()
	}

	wg.Wait()

	// Should not panic and should have a binding
	result, err := c.Make("key")
	if err != nil {
		t.Fatalf("Expected no error after concurrent operations, got %v", err)
	}

	// Result should be an integer (from one of the bindings)
	if _, ok := result.(int); !ok {
		t.Errorf("Expected int result, got %T", result)
	}
}

// TestInstance_NilValue tests that nil values can be stored
func TestInstance_NilValue(t *testing.T) {
	c := container.NewContainer()

	c.Instance("nil", nil)

	result, err := c.Make("nil")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// TestMake_PanicRecovery tests that panics in resolvers are caught and returned as errors
func TestMake_PanicRecovery(t *testing.T) {
	c := container.NewContainer()

	// Bind a resolver that panics
	c.Bind("panicky", func() interface{} {
		panic("resolver panic")
	})

	instance, err := c.Make("panicky")

	// Should not crash, should return error
	if instance != nil {
		t.Errorf("Expected nil instance after panic, got %v", instance)
	}

	if err == nil {
		t.Fatal("Expected error after panic, got nil")
	}

	// Error message should contain key and panic message
	expectedSubstrings := []string{"panic while resolving", "panicky", "resolver panic"}
	for _, substr := range expectedSubstrings {
		if !contains(err.Error(), substr) {
			t.Errorf("Expected error to contain '%s', got: %s", substr, err.Error())
		}
	}
}

// TestMake_PanicWithDifferentTypes tests panic recovery with different panic types
func TestMake_PanicWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name        string
		panicValue  interface{}
		expectedErr string
	}{
		{"string panic", "string error", "string error"},
		{"int panic", 42, "42"},
		{"struct panic", struct{ Msg string }{"struct error"}, "{struct error}"},
		{"nil panic", nil, "panic called with nil argument"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := container.NewContainer()

			c.Bind("test", func() interface{} {
				panic(tt.panicValue)
			})

			instance, err := c.Make("test")

			if instance != nil {
				t.Errorf("Expected nil instance, got %v", instance)
			}

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error to contain '%s', got: %s", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestMake_PanicInSingleton tests that panics in singleton resolvers are handled
func TestMake_PanicInSingleton(t *testing.T) {
	c := container.NewContainer()

	c.Singleton("panicky_singleton", func() interface{} {
		panic("singleton panic")
	})

	// First call should catch panic
	instance1, err1 := c.Make("panicky_singleton")
	if instance1 != nil {
		t.Errorf("Expected nil instance, got %v", instance1)
	}
	if err1 == nil {
		t.Fatal("Expected error, got nil")
	}

	// Second call should also fail (panic prevented caching)
	instance2, err2 := c.Make("panicky_singleton")
	if instance2 != nil {
		t.Errorf("Expected nil instance on second call, got %v", instance2)
	}
	if err2 == nil {
		t.Fatal("Expected error on second call, got nil")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
