package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/HewlettPackard/galadriel/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "conf/server/server.conf"

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig(cmd)
			if err != nil {
				return err
			}

			s := server.New(config)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			err = s.Run(ctx)
			if err != nil {
				return err
			}

			config.Logger.Info("Server stopped gracefully")
			return nil
		},
	}
}

func LoadConfig(cmd *cobra.Command) (*server.Config, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("cannot read flag config: %w", err)
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

	sc, err := NewServerConfig(c)
	if err != nil {
		return nil, fmt.Errorf("failed to build server configuration: %w", err)
	}

	logLevel, err := logrus.ParseLevel(c.Server.LogLevel)
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(logLevel)

	return sc, nil
}

func init() {
	runCmd := NewRunCmd()
	runCmd.Flags().StringP("config", "c", defaultConfigPath, "config file path")

	RootCmd.AddCommand(runCmd)
}
