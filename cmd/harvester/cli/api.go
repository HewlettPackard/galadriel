package cli

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "conf/harvester/harvester.conf"

var configPath string

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			HarvesterCLI.runHarvesterAPI(configPath)
		},
	}
}

func (c *HarvesterCli) runHarvesterAPI(configPath string) {
	c.logger.Info("Confinguring Harvester Cli")
	cfg, err := config.LoadFromDisk(configPath)
	if err != nil {
		c.logger.Error("Error loading config:", err)
	}

	ctx := context.Background()
	harvester.NewHarvesterManager().Start(ctx, *cfg)
}

func init() {
	runCmd := NewRunCmd()
	runCmd.PersistentFlags().StringVar(&configPath, "config", defaultConfigPath, "config file (default is conf/harvester/harvester.conf)")
	RootCmd.AddCommand(runCmd)
}
