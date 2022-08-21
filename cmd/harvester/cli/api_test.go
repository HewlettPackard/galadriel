package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRunCmd(t *testing.T) {
	configPath := ""
	expected := &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			HarvesterCLI.runHarvesterAPI(configPath)
		},
	}
	assert.ObjectsAreEqual(expected, NewRunCmd())
}
