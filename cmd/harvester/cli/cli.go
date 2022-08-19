package cli

import (
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
)

type HarvesterCli struct {
	logger *common.Logger
}

var HarvesterCLI *HarvesterCli

func NewHarvesterCli() *HarvesterCli {
	return &HarvesterCli{
		logger: common.NewLogger(telemetry.Harvester),
	}
}

func Run() int {
	HarvesterCLI = NewHarvesterCli()
	return HarvesterCLI.Run()
}

func (c *HarvesterCli) Run() int {
	c.logger.Info("Starting the Harvester CLI")

	return Execute()
}
