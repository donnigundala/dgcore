package foundation

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/donnigundala/dgcore/container"
	contractContainer "github.com/donnigundala/dgcore/contracts/container"
	"github.com/donnigundala/dgcore/contracts/foundation"
)

// Application is the central struct of the framework.
// It embeds the Container to provide dependency injection capabilities.
type Application struct {
	contractContainer.Container

	basePath  string
	providers []ServiceProvider
	booted    bool
	shutdown  *shutdownManager
}

// New creates a new Application instance.
func New(basePath string) *Application {
	app := &Application{
		Container: container.NewContainer(),
		basePath:  basePath,
		shutdown:  newShutdownManager(),
	}

	// Bind the application instance to the container
	app.Instance("app", app)
	app.Instance("container", app.Container)

	return app
}

// Register registers a service provider with the application.
func (app *Application) Register(provider foundation.ServiceProvider) {
	app.providers = append(app.providers, provider)
	provider.Register()

	if app.booted {
		provider.Boot()
	}
}

// Boot boots the application and all registered providers.
// Boot boots the application and all registered providers.
func (app *Application) Boot() {
	if app.booted {
		return
	}

	for _, provider := range app.providers {
		provider.Boot()
	}

	app.booted = true
}

// IsBooted checks if the application has been booted.
func (app *Application) IsBooted() bool {
	return app.booted
}

// BasePath returns the base path of the application.
func (app *Application) BasePath() string {
	return app.basePath
}

// ConfigPath returns the configuration path.
func (app *Application) ConfigPath() string {
	return filepath.Join(app.basePath, "config")
}

// DatabasePath returns the database path.
func (app *Application) DatabasePath() string {
	return filepath.Join(app.basePath, "database")
}

// StoragePath returns the storage path.
func (app *Application) StoragePath() string {
	return filepath.Join(app.basePath, "storage")
}

// Environment returns the current environment (e.g., local, production).
func (app *Application) Environment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return "production"
	}
	return env
}

// IsProduction checks if the application is running in production mode.
func (app *Application) IsProduction() bool {
	return app.Environment() == "production"
}

// Log returns the application logger.
// This assumes a logger is bound to the container, or falls back to default.
func (app *Application) Log() *slog.Logger {
	logger, err := app.Make("log")
	if err != nil {
		return slog.Default()
	}
	return logger.(*slog.Logger)
}
