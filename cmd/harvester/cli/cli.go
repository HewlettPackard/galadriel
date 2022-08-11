package cli

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
)

const defaultConfPath = "conf/harvester/harvester.conf"

type HarvesterCLI struct {
	logger *common.Logger
}

func NewHarvesterCLI() *HarvesterCLI {
	return &HarvesterCLI{
		logger: common.NewLogger("harvester"),
	}
}

func (c *HarvesterCLI) Run(args []string) int {
	if len(args) != 1 {
		c.logger.Error("Unknown arguments", args)
		return 1
	}

	cfg, err := config.LoadFromDisk(defaultConfPath)

	if err != nil {
		c.logger.Error("Error loading config:", err)
		return 1
	}

	ctx := context.Background()
	if args[0] == "run" {
		harvester.NewHarvesterManager().Start(ctx, *cfg)
	}

	c.logger.Error("Unknown command:", args[0])
	return 1
}
