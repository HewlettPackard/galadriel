package cli

import (
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/spf13/cobra"
)

type HarvesterCli struct {
	logger *common.Logger
	cli    *cobra.Command
}

func NewHarvesterCli() *HarvesterCli {
	return &HarvesterCli{
		logger: common.NewLogger(telemetry.Harvester),
		cli: &cobra.Command{
			Use:   "harvester",
			Short: "Run Galadriel Harvester CLI",
			Long:  "Command to run the Galadriel Harvester CLI",
			Run: func(cmd *cobra.Command, args []string) {
				RunHarvesterAPI()
			},
		},
	}
}

func (c *HarvesterCli) Run() int {
	c.logger.Info("Starting the Harvester CLI")

	err := c.cli.Execute()
	if err != nil {
		c.logger.Error("Unable to execute cli", err)
		return 1
	}

	return 0
}

func (c *HarvesterCli) AddCommand(cmd *cobra.Command) {
	c.cli.AddCommand(cmd)
}
