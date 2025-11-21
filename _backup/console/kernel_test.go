package console_test

import (
	"testing"

	"github.com/donnigundala/dgcore/console"
	contractConsole "github.com/donnigundala/dgcore/contracts/console"
	"github.com/spf13/cobra"
)

// MockCommand is a mock implementation of console.Command
type MockCommand struct {
	Sig  string
	Desc string
	Ran  bool
}

func (m *MockCommand) Signature() string {
	return m.Sig
}

func (m *MockCommand) Description() string {
	return m.Desc
}

func (m *MockCommand) Configure(cmd *cobra.Command) {
	// No-op
}

func (m *MockCommand) Handle(cmd *cobra.Command, args []string) error {
	m.Ran = true
	return nil
}

func TestKernel_RegisterAndCall(t *testing.T) {
	kernel := console.NewKernel()

	mockCmd := &MockCommand{
		Sig:  "test:command",
		Desc: "A test command",
	}

	kernel.Register([]contractConsole.Command{mockCmd})

	// Verify command is registered in cobra
	found := false
	for _, cmd := range kernel.RootCmd().Commands() {
		if cmd.Use == "test:command" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Command 'test:command' was not registered in cobra")
	}

	// Test Call
	err := kernel.Call("test:command", []string{})
	if err != nil {
		t.Errorf("Kernel.Call failed: %v", err)
	}

	if !mockCmd.Ran {
		t.Error("Mock command did not run")
	}
}
