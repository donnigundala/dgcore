package foundation

// ServiceProvider is the interface for all service providers.
// Service providers are responsible for registering and booting services.
type ServiceProvider interface {
	// Register registers services into the application container.
	// This is called before Boot() and should only register bindings.
	Register(app Application) error

	// Boot bootstraps services after all providers have been registered.
	// This is where you can safely resolve dependencies from the container.
	Boot(app Application) error
}

// PluginProvider extends ServiceProvider with metadata for plugin discovery.
// This is optional - only implement if your provider is a plugin that needs metadata.
type PluginProvider interface {
	ServiceProvider

	// Name returns the plugin name (e.g., "websocket", "metrics")
	Name() string

	// Version returns the plugin version (e.g., "1.0.0")
	Version() string

	// Dependencies returns a list of plugin names this plugin depends on
	Dependencies() []string
}

// BeforeRegisterProvider is an optional interface for providers that need
// to perform actions before registration.
type BeforeRegisterProvider interface {
	BeforeRegister(app Application) error
}

// AfterBootProvider is an optional interface for providers that need
// to perform actions after boot.
type AfterBootProvider interface {
	AfterBoot(app Application) error
}

// ShutdownProvider is an optional interface for providers that need
// to perform cleanup during application shutdown.
type ShutdownProvider interface {
	Shutdown(app Application) error
}

// CommandProvider is an optional interface for providers that want to
// register console commands.
type CommandProvider interface {
	// Commands returns a list of console commands to register.
	// We use interface{} here to avoid circular dependencies with the console package,
	// but the expected type is []console.Command.
	Commands() []interface{}
}
