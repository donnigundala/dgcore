package foundation

import (
	"github.com/donnigundala/dg-core/contracts/container"
)

// Application is the interface for the application.
type Application interface {
	container.Container

	// Lifecycle
	Boot() error
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
	Register(provider ServiceProvider) error
	GetProviders() []ServiceProvider
	HasProvider(name string) bool
}
