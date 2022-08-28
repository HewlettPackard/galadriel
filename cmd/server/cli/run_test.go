package cli

import (
	"errors"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/server/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRunCmd(t *testing.T) {
	configPath = ""
	expected := &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}

			runServerFn(configPath)
			return nil
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

func TestRunServerAPI(t *testing.T) {
	runAPIcalled := false

	runAPI = func() {
		runAPIcalled = true
	}

	loadConfigFromDisk = func(path string) (*config.Server, error) {
		return nil, errors.New("ops")
	}

	ServerCLI.runServerAPI(defaultConfigPath)

	assert.Equal(t, false, runAPIcalled)

	loadConfigFromDisk = func(path string) (*config.Server, error) {
		return nil, nil
	}

	ServerCLI.runServerAPI(defaultConfigPath)

	assert.Equal(t, true, runAPIcalled)
}
