package container

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
