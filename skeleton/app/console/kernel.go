package console

import (
	"github.com/donnigundala/dgcore/console"
	contractConsole "github.com/donnigundala/dgcore/contracts/console"
	"github.com/donnigundala/dgcore/foundation"

	"example-app/app/console/commands"
)

// NewKernel creates a new Console Kernel with user commands registered.
func NewKernel(app *foundation.Application) contractConsole.Kernel {
	k := console.NewKernel()

	// Register user commands here
	k.Register([]contractConsole.Command{
		&commands.ServeCommand{App: app},
	})

	return k
}
