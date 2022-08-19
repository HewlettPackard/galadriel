package cli

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:  "harvester",
	Long: "This is Galadriel Harvester CLI",
}

var cmdExecute = RootCmd.Execute

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}
	return 0
}
