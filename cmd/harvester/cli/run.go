package cli

import (
	"context"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

const defaultConfigPath = "conf/harvester/harvester.conf"

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel Harvester",
		Long:  "Run this command to start the Galadriel Harvester",
		RunE: func(cmd *cobra.Command, args []string) error {

			config, err := LoadConfig(cmd)
			if err != nil {
				return err
			}

			h := harvester.New(config)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			err = h.Run(ctx)
			if err != nil {
				return err
			}

			config.Log.Info("Harvester stopped gracefully")
			return nil
		},
	}
}

func LoadConfig(cmd *cobra.Command) (*harvester.Config, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("cannot get flag config: %w", err)
	}

	if configPath == "" {
		configPath = defaultConfigPath
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open configuration file: %w", err)
	}
	defer configFile.Close()

	c, err := ParseConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	hc, err := NewHarvesterConfig(c)
	if err != nil {
		return nil, fmt.Errorf("failed to build harvester configuration: %w", err)
	}
	return hc, nil
}

func init() {
	runCmd := NewRunCmd()
	runCmd.PersistentFlags().StringP("config", "c", defaultConfigPath, "Config file")
	RootCmd.AddCommand(runCmd)
}
