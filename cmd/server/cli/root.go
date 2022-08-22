package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "server",
		Long: "This is Galadriel Server CLI",
	}
}

var RootCmd = &cobra.Command{}

var cmdExecute = RootCmd.Execute

func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}

	return 0
}
