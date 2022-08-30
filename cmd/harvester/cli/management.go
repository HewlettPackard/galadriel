package cli

import (
	"github.com/spf13/cobra"
)

func NewFederationtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "federation",
		Short: "Manage federation relationships",
		Long:  "Run this command to approve and deny relationships",
		Run: func(cmd *cobra.Command, args []string) {
			HarvesterCLI.runManagementAPI()
		},
	}
}

func (c *HarvesterCli) runManagementAPI() {
	c.logger.Info("Starting Management API")
}

func init() {
	runCmd := NewFederationtCmd()
	RootCmd.AddCommand(runCmd)
}
