package health

import (
	"context"
	"fmt"
)

// simpleCheck is a basic health check implementation.
type simpleCheck struct {
	name string
	fn   func(context.Context) error
}

func (c *simpleCheck) Check(ctx context.Context) error {
	return c.fn(ctx)
}

func (c *simpleCheck) Name() string {
	return c.name
}

// SimpleCheck creates a simple health check with a custom function.
func SimpleCheck(name string, fn func(context.Context) error) Checker {
	return &simpleCheck{
		name: name,
		fn:   fn,
	}
}

// AlwaysHealthy returns a health check that always passes.
func AlwaysHealthy(name string) Checker {
	return SimpleCheck(name, func(ctx context.Context) error {
		return nil
	})
}

// DatabaseCheck creates a health check for database connectivity.
// The ping function should attempt to ping the database.
func DatabaseCheck(name string, ping func(context.Context) error) Checker {
	return SimpleCheck(fmt.Sprintf("database:%s", name), ping)
}

// CacheCheck creates a health check for cache connectivity.
// The ping function should attempt to ping the cache.
func CacheCheck(name string, ping func(context.Context) error) Checker {
	return SimpleCheck(fmt.Sprintf("cache:%s", name), ping)
}

// CustomCheck creates a custom health check.
func CustomCheck(name string, fn func(context.Context) error) Checker {
	return SimpleCheck(name, fn)
}
