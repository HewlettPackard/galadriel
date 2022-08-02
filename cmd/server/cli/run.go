package cli

import (
	"github.com/HewlettPackard/Galadriel/cmd/server/api"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start HTTP Galadriel server",
	Long: `Run this command to start HTTP Galadriel server`,
	Run: func(cmd *cobra.Command, args []string) {
		api.Run()
	},
}

func init() {
	serverCmd.AddCommand(runCmd)
}
