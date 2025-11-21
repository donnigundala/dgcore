package console

import (
	"github.com/donnigundala/dgcore/console/db"
	"github.com/donnigundala/dgcore/console/migrate"
	"github.com/donnigundala/dgcore/contracts/console"
	"github.com/spf13/cobra"
)

// Kernel is the concrete implementation of the Console kernel.
type Kernel struct {
	rootCmd *cobra.Command
}

// NewKernel creates a new Kernel instance.
func NewKernel() *Kernel {
	k := &Kernel{
		rootCmd: &cobra.Command{
			Use:   "dg",
			Short: "DG Framework Console",
			Long:  `The DG Framework Console Application.`,
		},
	}

	// Register default commands
	k.Register([]console.Command{
		&db.WipeCommand{},
		&migrate.MakeMigrationCommand{},
	})

	return k
}

// Handle handles the incoming console command execution.
func (k *Kernel) Handle() error {
	return k.rootCmd.Execute()
}

// Register registers commands with the kernel.
func (k *Kernel) Register(commands []console.Command) {
	for _, cmd := range commands {
		k.rootCmd.AddCommand(ToCobra(cmd))
	}
}

// Call calls a specific command by name.
func (k *Kernel) Call(command string, args []string) error {
	k.rootCmd.SetArgs(append([]string{command}, args...))
	return k.rootCmd.Execute()
}

// RootCmd returns the underlying cobra root command.
func (k *Kernel) RootCmd() *cobra.Command {
	return k.rootCmd
}
