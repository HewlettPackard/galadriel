package server

import (
	"github.com/HewlettPackard/galadriel/cmd/server/cli"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Commands to manage Galadriel server",
	Long:  "Commands to manage Galadriel server",
}

func init() {
	cli.RootCmd.AddCommand(ServerCmd)
}
