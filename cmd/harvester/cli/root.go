package cli

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:  "harvester",
	Long: "This is Galadriel Harvester CLI",
}

var cmdExecute = RootCmd.Execute

func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}
	return 0
}
