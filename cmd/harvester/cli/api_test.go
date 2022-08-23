package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRunCmd(t *testing.T) {
	expected := &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}

			err = HarvesterCLI.runHarvesterAPI(configPath)
			if err != nil {
				return err
			}

			return nil
		},
	}
	assert.ObjectsAreEqual(expected, NewRunCmd())
}
