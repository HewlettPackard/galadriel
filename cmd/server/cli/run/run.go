package run

import (
	"github.com/HewlettPackard/galadriel/cmd/server/api"
	"github.com/HewlettPackard/galadriel/cmd/server/cli/server"
	"github.com/spf13/cobra"
)

var runServerFn = api.Run

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			runServerFn()
		},
	}
}

func init() {
	server.ServerCmd.AddCommand(NewRunCmd())
}
