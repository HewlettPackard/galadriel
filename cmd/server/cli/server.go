package cli

import (
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Commands to manage Galadriel server",
	Long:  "Commands to manage Galadriel server",
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
