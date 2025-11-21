package foundation

// ServiceProvider is the interface for all service providers.
type ServiceProvider interface {
	// Register registers the service provider.
	Register()
	// Boot boots the service provider.
	Boot()
}
