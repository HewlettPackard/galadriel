package cli

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{}
var cmdExecute = RootCmd.Execute

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "harvester",
		Long: "This is Galadriel Harvester CLI",
	}
}

func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}
	return 0
}
