package cli

import (
	"errors"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/spf13/cobra"
)

var managementEntity string
var action string
var id string

var defaultManagementEntity = telemetry.Federation
var defaultAction = telemetry.List

func NewManagementCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manages member and federation relationships",
		Long:  "Run this command to approve and deny relationships between members or federations",
		RunE: func(cmd *cobra.Command, args []string) error {
			managementObject, err := cmd.Flags().GetString("object")
			if err != nil {
				return err
			}

			action, err := cmd.Flags().GetString("action")
			if err != nil {
				return err
			}

			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			HarvesterCLI.runManagementAPI(managementObject, action, id)
			return nil
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
	runCmd.PersistentFlags().StringVar(&managementEntity, "entity", defaultManagementEntity, "choose what object to manage between federation and member")
	runCmd.PersistentFlags().StringVar(&action, "action", defaultAction, "choose what action to do with the object selected")
	runCmd.PersistentFlags().StringVar(&id, "id", "", "choose what id will be acted on")

	RootCmd.AddCommand(runCmd)
}
