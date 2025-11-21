package main

import (
	"fmt"
	"os"

	"github.com/donnigundala/dgcore/contracts/console"
	"github.com/donnigundala/dgcore/foundation"
)

func main() {
	// Initialize the application
	app := foundation.New("/tmp")

	// Resolve the Console Kernel
	kernelInstance, err := app.Make("console.kernel")
	if err != nil {
		fmt.Printf("Error resolving console kernel: %v\n", err)
		os.Exit(1)
	}
	kernel := kernelInstance.(console.Kernel)

	// Handle the command
	if err := kernel.Handle(); err != nil {
		os.Exit(1)
	}
}
