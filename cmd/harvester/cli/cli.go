package cli

import (
	"github.com/spf13/cobra"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
)

var (
	RootCmd      = NewRootCmd()
	cmdExecute   = RootCmd.Execute
	HarvesterCLI = &HarvesterCli{
		logger: common.NewLogger(telemetry.Harvester),
	}
)

type HarvesterCli struct {
	logger *common.Logger
}

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "harvester",
		Long: "This is the Galadriel Harvester CLI",
	}
}

func Run() int {
	err := HarvesterCLI.Run()
	if err != nil {
		return 1
	}

	return 0
}

func (c *HarvesterCli) Run() error {
	c.logger.Info("Starting the Harvester CLI")
	return c.Execute()
}

func (c *HarvesterCli) Execute() error {
	err := cmdExecute()
	if err != nil {
		c.logger.Error("Error when executing the Harvester CLI:", err)
		return err
	}

	return nil
}
