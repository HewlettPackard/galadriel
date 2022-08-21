package cli

import (
	"errors"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/spf13/cobra"
)

var managementObject string
var action string
var id string

var defaultManagementObject = telemetry.Federation
var defaultAction = telemetry.List

func NewManagementCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage member and federation relationships",
		Long:  "Run this command to approve and deny relationships between members or federations",
		Run: func(cmd *cobra.Command, args []string) {
			managementObject, _ := cmd.Flags().GetString("object")
			action, _ := cmd.Flags().GetString("action")
			id, _ := cmd.Flags().GetString("id")
			HarvesterCLI.runManagementAPI(managementObject, action, id)
		},
	}
}

func (c *HarvesterCli) runManagementAPI(managementObject, action, id string) {
	c.logger.Info("Starting Management API")

	if managementObject == telemetry.Federation {
		err := c.runFederationAction(action, id)
		if err != nil {
			c.logger.Error(err)
		}
	}
}

func (c *HarvesterCli) runFederationAction(action, id string) error {
	if action == telemetry.List {
		c.logger.Info(telemetry.Federation, action, id)
		return nil
	}

	if action == telemetry.Approve {
		c.logger.Info(telemetry.Federation, action, id)
		return nil
	}

	if action == telemetry.Deny {
		c.logger.Info(telemetry.Federation, action, id)
		return nil
	}

	return errors.New("action not found")
}

func init() {
	runCmd := NewManagementCmd()
	runCmd.PersistentFlags().StringVar(&managementObject, "object", defaultManagementObject, "choose what object to manage between federation and member")
	runCmd.PersistentFlags().StringVar(&action, "action", defaultAction, "choose what action to do with the object selected")
	runCmd.PersistentFlags().StringVar(&id, "id", "", "choose what id will be acted on")

	RootCmd.AddCommand(runCmd)
}
