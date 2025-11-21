package console

import "github.com/spf13/cobra"

// Command defines the interface for a console command.
type Command interface {
	// Signature returns the command name/signature (e.g. "migrate:up").
	Signature() string

	// Description returns the short description of the command.
	Description() string

	// Handle executes the command logic.
	Handle(cmd *cobra.Command, args []string) error

	// Configure configures the command (e.g. flags, arguments).
	Configure(cmd *cobra.Command)
}
