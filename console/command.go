package console

import (
	"github.com/donnigundala/dgcore/contracts/console"
	"github.com/spf13/cobra"
)

// BaseCommand is a helper struct that implements the Command interface.
// Users can embed this struct to get default implementations if needed,
// but they should implement the interface methods.
type BaseCommand struct{}

// ToCobra converts a framework Command to a cobra.Command.
func ToCobra(cmd console.Command) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   cmd.Signature(),
		Short: cmd.Description(),
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.Handle(c, args)
		},
	}
	cmd.Configure(cobraCmd)
	return cobraCmd
}
