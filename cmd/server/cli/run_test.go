package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TODO: this test does not seem to add any value
func TestNewRunCmd(t *testing.T) {
	configPath = ""
	expected := &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	assert.ObjectsAreEqual(expected, NewRunCmd())
}

func TestRunCommand(t *testing.T) {
	called := false
	runServerFn = func(config string) {
		called = true
	}

	cmd := NewRunCmd()
	cmd.Flags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file path")
	err := cmd.Execute()

	assert.Equal(t, nil, err, "unexpected error")
	assert.Equal(t, true, called, "failed to call runServerFn")
}
