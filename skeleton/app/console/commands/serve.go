package commands

import (
	"github.com/donnigundala/dgcore/console"
	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/foundation"
	coreHTTP "github.com/donnigundala/dgcore/http"
	"github.com/spf13/cobra"
)

type ServeCommand struct {
	console.BaseCommand
	App *foundation.Application
}

func (c *ServeCommand) Signature() string {
	return "serve"
}

func (c *ServeCommand) Description() string {
	return "Serve the application on the Go HTTP server"
}

func (c *ServeCommand) Handle(cmd *cobra.Command, args []string) error {
	// Resolve Kernel
	kernelInstance, err := c.App.Make("kernel")
	if err != nil {
		return err
	}
	kernel := kernelInstance.(contractHTTP.Kernel)

	// Create Config (hardcoded for now)
	cfg := coreHTTP.Config{
		Addr: ":8080",
	}

	// Create Server
	server := coreHTTP.NewHTTPServer(cfg, kernel)

	// Start Server
	return server.Start()
}

func (c *ServeCommand) Configure(cmd *cobra.Command) {
	// Add flags if needed
}
