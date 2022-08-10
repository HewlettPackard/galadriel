package cli

import (
	"github.com/HewlettPackard/Galadriel/cmd/server/api"
	"github.com/spf13/cobra"
)

var runServerFn = api.Run

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Start HTTP Galadriel server",
		Long:  "Run this command to start HTTP Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			runServerFn()
		},
	}
}

func init() {
	serverCmd.AddCommand(NewRunCmd())
}
