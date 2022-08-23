package cli

import (
	"github.com/HewlettPackard/galadriel/cmd/server/api"
	"github.com/HewlettPackard/galadriel/pkg/server/config"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "conf/server/server.conf"

var configPath string

var runServerFn = api.Run

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			ServerCLI.runServerAPI(configPath)
		},
	}
}

func (c *serverCLI) runServerAPI(configPath string) {
	c.logger.Info("Configuring Harvester CLI")

	// TODO: pass config variables to runServerFn()
	_, err := config.LoadFromDisk(configPath)
	if err != nil {
		c.logger.Error("Error loading config:", err)
	}

	runServerFn()
}

func init() {
	runCmd := NewRunCmd()
	runCmd.Flags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file path")

	RootCmd.AddCommand(runCmd)
}
