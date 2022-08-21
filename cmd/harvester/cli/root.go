package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "harvester",
		Long: "This is Galadriel Harvester CLI",
	}
}

var RootCmd = NewRootCmd()

var cmdExecute = RootCmd.Execute

func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}

	return 0
}
