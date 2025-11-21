package foundation

import (
	"github.com/donnigundala/dgcore/contracts/container"
)

// Application is the interface for the application.
type Application interface {
	container.Container

	// Lifecycle
	Boot()
	IsBooted() bool

	// Paths
	BasePath() string
	ConfigPath() string
	DatabasePath() string
	StoragePath() string

	// Environment
	Environment() string
	IsProduction() bool

	// Service Providers
	Register(provider ServiceProvider)
}
