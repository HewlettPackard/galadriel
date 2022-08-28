package cli

import (
	"github.com/HewlettPackard/galadriel/pkg/server"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "conf/server/server.conf"

var configPath string
var runServerFn = server.Run

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}

			runServerFn(configPath)
			return nil
		},
	}
}

func init() {
	runCmd := NewRunCmd()
	runCmd.Flags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file path")

	RootCmd.AddCommand(runCmd)
}
