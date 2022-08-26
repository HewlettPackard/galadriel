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
		Short: "Runs the Galadriel Harvester",
		Long:  "Run this command to start the Galadriel Harvester",
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
}

func (c *HarvesterCli) runHarvesterAPI(configPath string) error {
	c.logger.Info("Configuring Harvester Cli")

	cfg, err := config.LoadFromDisk(configPath)
	if err != nil {
		return err
	}

	ctx := context.Background()
	harvester.NewHarvesterManager().Start(ctx, *cfg)

	return nil
}

func init() {
	runCmd := NewRunCmd()
	runCmd.PersistentFlags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file")

	RootCmd.AddCommand(runCmd)
}
