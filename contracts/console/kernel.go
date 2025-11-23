package console

// Kernel defines the interface for the Console kernel.
type Kernel interface {
	// Handle handles the incoming console command execution.
	Handle() error

	// Register registers commands with the kernel.
	Register(commands []Command)

	// Call calls a specific command by name.
	Call(command string, args []string) error
}
