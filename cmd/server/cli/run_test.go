package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRunCmd(t *testing.T) {
	configPath = ""
	expected := &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			ServerCLI.runServerAPI(configPath)
		},
	}
	assert.ObjectsAreEqual(expected, NewRunCmd())
}

func TestRunCommand(t *testing.T) {
	called := false

	runServerFn = func() {
		called = true
	}

	cmd := NewRunCmd()
	err := cmd.Execute()

	assert.Equal(t, err, nil)
	assert.Equal(t, called, true)
}
