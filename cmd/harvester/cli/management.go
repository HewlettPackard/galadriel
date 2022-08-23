package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func NewManagementCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manages member and federation relationships",
		Long:  "Run this command to approve and deny relationships between members or federations",
		RunE: func(cmd *cobra.Command, args []string) error {
			HarvesterCLI.runManagementAPI()
			return errors.New("not implemented")
		},
	}
}

func (c *HarvesterCli) runManagementAPI() {
	c.logger.Info("Starting Management API")
}

func init() {
	runCmd := NewManagementCmd()
	RootCmd.AddCommand(runCmd)
}
