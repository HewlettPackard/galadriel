package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRunCmd(t *testing.T) {
	expected := &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel Harvester",
		Long:  "Run this command to start the Galadriel Harvester",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")

			HarvesterCLI.runHarvesterAPI(configPath)
		},
	}
	assert.ObjectsAreEqual(expected, NewRunCmd())
}
