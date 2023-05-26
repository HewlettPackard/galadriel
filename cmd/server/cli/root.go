package cli

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{}
var cmdExecute = RootCmd.Execute

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "server",
		Long: "This is Galadriel Server CLI",
	}
}

func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}
	return 0
}

func init() {
	RootCmd.PersistentFlags().StringP("socketPath", "", defaultSocketPath, "Path to the Galadriel Server API socket")
}
