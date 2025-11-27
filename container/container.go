// Package container provides a dependency injection container for managing application dependencies.
//
// The container supports both transient and singleton bindings, allowing you to register
// and resolve dependencies throughout your application. It is thread-safe and can be used
// concurrently from multiple goroutines.
//
// # Basic Usage
//
// Create a container and register bindings:
//
//	c := container.NewContainer()
//
//	// Register a transient binding (new instance each time)
//	c.Bind("logger", func() interface{} {
//	    return &Logger{Level: "info"}
//	})
//
//	// Register a singleton (shared instance)
//	c.Singleton("db", func() interface{} {
//	    return &Database{Host: "localhost"}
//	})
//
//	// Register an existing instance
//	config := &Config{Port: 8080}
//	c.Instance("config", config)
//
// Resolve dependencies:
//
//	db, err := c.Make("db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	database := db.(*Database)
//
// # Thread Safety
//
// All container operations are thread-safe and can be called concurrently:
//
//	var wg sync.WaitGroup
//	for i := 0; i < 100; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        db, _ := c.Make("db") // Safe concurrent access
//	        // Use db...
//	    }()
//	}
//	wg.Wait()
//
// # Dependency Injection
//
// Resolvers can accept the container as a parameter for dependency injection:
//
//	c.Singleton("service", func(c container.Container) interface{} {
//	    db, _ := c.Make("db")
//	    logger, _ := c.Make("logger")
//	    return &Service{
//	        DB:     db.(*Database),
//	        Logger: logger.(*Logger),
//	    }
//	})
package container

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/donnigundala/dg-core/contracts/container"
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

// NewContainer creates a new dependency injection container instance.
//
// The container is initialized with empty binding and instance maps, ready to
// register and resolve dependencies. All operations on the returned container
// are thread-safe.
//
// Example:
//
//	c := container.NewContainer()
//
//	// Register bindings
//	c.Singleton("db", func() interface{} {
//	    return &Database{Host: "localhost", Port: 5432}
//	})
//
//	// Resolve dependencies
//	db, err := c.Make("db")
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewContainer() container.Container {
	return &containerImpl{
		bindings:  make(map[string]binding),
		instances: make(map[string]interface{}),
	}
}

// Bind registers a transient binding with the container.
//
// Transient bindings create a new instance every time Make is called.
// The resolver should be a function that returns the dependency instance.
//
// The resolver function can optionally accept the container as a parameter
// for dependency injection:
//
//	// Simple resolver
//	c.Bind("logger", func() interface{} {
//	    return &Logger{Level: "debug"}
//	})
//
//	// Resolver with dependency injection
//	c.Bind("service", func(c container.Container) interface{} {
//	    logger, _ := c.Make("logger")
//	    return &Service{Logger: logger.(*Logger)}
//	})
func (c *containerImpl) Bind(key string, resolver interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings[key] = binding{
		resolver: resolver,
		shared:   false,
	}
}

// Singleton registers a shared binding in the container.
//
// Singleton bindings create only one instance, which is cached and reused
// for all subsequent Make calls. This is useful for shared resources like
// database connections, configuration, or loggers.
//
// Example:
//
//	c.Singleton("db", func() interface{} {
//	    db, _ := sql.Open("postgres", "connection-string")
//	    return db
//	})
//
//	// First call creates the instance
//	db1, _ := c.Make("db")
//
//	// Subsequent calls return the same instance
//	db2, _ := c.Make("db")
//	// db1 == db2 (same pointer)
func (c *containerImpl) Singleton(key string, resolver interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings[key] = binding{
		resolver: resolver,
		shared:   true,
	}
}

// Instance registers an existing instance as shared in the container.
//
// This is useful when you already have an instance and want to register it
// directly without a resolver function. The instance will be returned as-is
// for all Make calls.
//
// Example:
//
//	config := &Config{
//	    Port:     8080,
//	    Host:     "localhost",
//	    LogLevel: "info",
//	}
//	c.Instance("config", config)
//
//	// Retrieve the same instance
//	cfg, _ := c.Make("config")
//	// cfg == config (same pointer)
func (c *containerImpl) Instance(key string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.instances[key] = instance
}

// Make resolves the given type from the container.
//
// Make first checks if an instance already exists (for singletons or instances
// registered via Instance). If not, it calls the resolver function to create
// a new instance. For singleton bindings, the created instance is cached for
// future calls.
//
// The resolver function can optionally accept the container as its first parameter,
// enabling dependency injection:
//
//	c.Singleton("service", func(c container.Container) interface{} {
//	    db, _ := c.Make("db")
//	    logger, _ := c.Make("logger")
//	    return &Service{DB: db, Logger: logger}
//	})
//
// Make returns an error if the binding is not found:
//
//	instance, err := c.Make("unknown")
//	if err != nil {
//	    // Handle error: binding not found
//	}
//
// Panic Recovery: If a resolver function panics during execution, Make will
// recover from the panic and return it as an error instead of crashing the
// application. This ensures production safety:
//
//	c.Bind("bad", func() interface{} {
//	    panic("something went wrong")
//	})
//	instance, err := c.Make("bad")
//	// err will contain: "panic while resolving 'bad': something went wrong"
//
// Thread Safety: Make is safe for concurrent use. For singleton bindings,
// double-checked locking ensures only one instance is created even when
// called concurrently.
func (c *containerImpl) Make(key string) (instance interface{}, err error) {
	// Panic recovery - convert panics to errors for production safety
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while resolving '%s': %v", key, r)
			instance = nil
		}
	}()

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
//
// This is useful for testing or when you need to reset the container state.
// After calling Flush, all previously registered bindings and instances are
// removed, and the container is empty.
//
// Example:
//
//	c.Singleton("db", dbResolver)
//	c.Instance("config", config)
//
//	// Clear everything
//	c.Flush()
//
//	// Container is now empty
//	_, err := c.Make("db") // Returns error: binding not found
//
// Thread Safety: Flush is safe for concurrent use, but should typically
// be called when no other goroutines are accessing the container.
func (c *containerImpl) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings = make(map[string]binding)
	c.instances = make(map[string]interface{})
}
