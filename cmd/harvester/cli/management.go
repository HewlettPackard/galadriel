package cli

import (
	"github.com/spf13/cobra"
)

func NewFederationtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "federation",
		Short: "Manage federation relationships",
		Long:  "Run this command to approve and deny relationships",
	}
}

func init() {
	runCmd := NewFederationtCmd()
	RootCmd.AddCommand(runCmd)
}
