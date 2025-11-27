package foundation

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/donnigundala/dg-core/container"
	contractContainer "github.com/donnigundala/dg-core/contracts/container"
	"github.com/donnigundala/dg-core/contracts/foundation"
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

// Register registers a service provider.
func (app *Application) Register(provider foundation.ServiceProvider) error {
	// Auto-inject configuration if provider has config fields
	if err := InjectProviderConfig(provider); err != nil {
		return fmt.Errorf("config injection failed for provider: %w", err)
	}

	// Check for BeforeRegister hook
	if hook, ok := provider.(foundation.BeforeRegisterProvider); ok {
		if err := hook.BeforeRegister(app); err != nil {
			return fmt.Errorf("BeforeRegister hook failed: %w", err)
		}
	}

	// Register the provider
	if err := provider.Register(app); err != nil {
		return err
	}

	// Check for Shutdown hook
	if hook, ok := provider.(foundation.ShutdownProvider); ok {
		app.RegisterShutdownHook(func() {
			if err := hook.Shutdown(app); err != nil {
				// Log error if logger is available, otherwise just print to stderr
				if logger := app.Log(); logger != nil {
					logger.Error("Provider shutdown failed", "error", err)
				} else {
					fmt.Fprintf(os.Stderr, "Provider shutdown failed: %v\n", err)
				}
			}
		})
	}

	// Add to providers list only after successful registration
	app.providers = append(app.providers, provider)

	// If app is already booted, boot this provider immediately
	if app.booted {
		if err := provider.Boot(app); err != nil {
			return fmt.Errorf("failed to boot provider: %w", err)
		}

		// Check for AfterBoot hook (since we just booted)
		if hook, ok := provider.(foundation.AfterBootProvider); ok {
			if err := hook.AfterBoot(app); err != nil {
				return fmt.Errorf("AfterBoot hook failed: %w", err)
			}
		}
	}

	return nil
}

// Boot boots the application and all registered providers.
func (app *Application) Boot() error {
	if app.booted {
		return nil
	}

	for _, provider := range app.providers {
		if err := provider.Boot(app); err != nil {
			return fmt.Errorf("failed to boot provider: %w", err)
		}
	}

	// Run AfterBoot hooks
	for _, provider := range app.providers {
		if hook, ok := provider.(foundation.AfterBootProvider); ok {
			if err := hook.AfterBoot(app); err != nil {
				return fmt.Errorf("AfterBoot hook failed: %w", err)
			}
		}
	}

	app.booted = true
	return nil
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

// GetProviders returns all registered service providers.
func (app *Application) GetProviders() []foundation.ServiceProvider {
	// Return a copy to prevent external modification
	providers := make([]foundation.ServiceProvider, len(app.providers))
	copy(providers, app.providers)
	return providers
}

// HasPlugin checks if a plugin with the given name is registered.
// Only providers implementing PluginProvider (with Name/Version/Dependencies) are checked.
func (app *Application) HasPlugin(name string) bool {
	for _, p := range app.providers {
		if plugin, ok := p.(foundation.PluginProvider); ok {
			if plugin.Name() == name {
				return true
			}
		}
	}
	return false
}

// HasProvider is deprecated. Use HasPlugin instead.
// This method is kept for backward compatibility and will be removed in v2.0.
//
// Deprecated: Use HasPlugin instead.
func (app *Application) HasProvider(name string) bool {
	return app.HasPlugin(name)
}

// RegisterPlugin registers a plugin provider with validation.
// It prevents duplicate plugin registrations by checking if a plugin
// with the same name is already registered.
//
// Example:
//
//	plugin := NewSlowQueryPlugin(config)
//	if err := app.RegisterPlugin(plugin); err != nil {
//	    return fmt.Errorf("failed to register plugin: %w", err)
//	}
func (app *Application) RegisterPlugin(plugin foundation.PluginProvider) error {
	// Validate: prevent duplicate registration
	if app.HasPlugin(plugin.Name()) {
		return fmt.Errorf("plugin '%s' is already registered", plugin.Name())
	}

	// Register as a normal service provider
	return app.Register(plugin)
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
