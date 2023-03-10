package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/HewlettPackard/galadriel/pkg/gca"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "conf/gca/gca.conf"

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel CA",
		Long:  "Run this command to start the Galadriel CA",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("cannot read flag config: %w", err)
			}
			if configPath == "" {
				configPath = defaultConfigPath
			}

			config, err := LoadConfig(configPath)
			if err != nil {
				return err
			}

			s, err := gca.NewGCA(config)
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			err = s.Run(ctx)
			if err != nil {
				return err
			}

			config.Logger.Info("GCA stopped gracefully")
			return nil
		},
	}
}

func LoadConfig(configPath string) (*gca.Config, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open configuration file: %w", err)
	}
	defer configFile.Close()

	c, err := ParseConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	sc, err := NewGCAConfig(c)
	if err != nil {
		return nil, fmt.Errorf("failed to build server configuration: %w", err)
	}

	return sc, nil
}

func init() {
	runCmd := NewRunCmd()
	runCmd.Flags().StringP("config", "c", defaultConfigPath, "config file path")

	RootCmd.AddCommand(runCmd)
}
