package main

import (
	"github.com/donnigundala/dgcore/cmd"
	// Import other packages as needed for CLI commands
	_ "github.com/donnigundala/dgcore/cmd/migrate" // Blank import to register migrate commands
)

func main() {
	cmd.Execute()
}
