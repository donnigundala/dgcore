package container

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/donnigundala/dgcore/contracts/container"
)

// Container is the interface for the dependency injection container.
type Container interface {
	// Bind registers a binding with the container.
	Bind(key string, resolver interface{})
	// Singleton registers a shared binding in the container.
	Singleton(key string, resolver interface{})
	// Instance registers an existing instance as shared in the container.
	Instance(key string, instance interface{})
	// Make resolves the given type from the container.
	Make(key string) (interface{}, error)
	// Flush removes all bindings and instances from the container.
	Flush()
}

// binding represents a registered dependency.
type binding struct {
	resolver interface{} // The function that resolves the dependency
	shared   bool        // Whether the instance should be shared (singleton)
}

// container is the concrete implementation of the Container interface.
type containerImpl struct {
	mu        sync.RWMutex
	bindings  map[string]binding
	instances map[string]interface{}
}

// NewContainer creates a new container instance.
func NewContainer() container.Container {
	return &containerImpl{
		bindings:  make(map[string]binding),
		instances: make(map[string]interface{}),
	}
}

// Bind registers a binding with the container.
func (c *containerImpl) Bind(key string, resolver interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings[key] = binding{
		resolver: resolver,
		shared:   false,
	}
}

// Singleton registers a shared binding in the container.
func (c *containerImpl) Singleton(key string, resolver interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings[key] = binding{
		resolver: resolver,
		shared:   true,
	}
}

// Instance registers an existing instance as shared in the container.
func (c *containerImpl) Instance(key string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.instances[key] = instance
}

// Make resolves the given type from the container.
func (c *containerImpl) Make(key string) (interface{}, error) {
	c.mu.RLock()
	if instance, ok := c.instances[key]; ok {
		c.mu.RUnlock()
		return instance, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check locking
	if instance, ok := c.instances[key]; ok {
		return instance, nil
	}

	binding, ok := c.bindings[key]
	if !ok {
		return nil, fmt.Errorf("binding not found for key: %s", key)
	}

	resolverVal := reflect.ValueOf(binding.resolver)
	if resolverVal.Kind() == reflect.Func {
		// For simplicity, we assume the resolver function takes no arguments
		// or we can implement dependency injection here later.
		// For now, let's assume it takes a Container as argument if it has 1 arg.
		args := []reflect.Value{}
		if resolverVal.Type().NumIn() == 1 {
			// Pass the container itself if the resolver expects it
			// We need to cast *containerImpl to container.Container
			args = append(args, reflect.ValueOf(c))
		}

		results := resolverVal.Call(args)
		if len(results) == 0 {
			return nil, fmt.Errorf("resolver returned no results for key: %s", key)
		}
		instance := results[0].Interface()

		if binding.shared {
			c.instances[key] = instance
		}
		return instance, nil
	}

	// If resolver is not a function, return it as is (though Bind usually expects a function)
	return binding.resolver, nil
}

// Flush removes all bindings and instances from the container.
func (c *containerImpl) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings = make(map[string]binding)
	c.instances = make(map[string]interface{})
}
