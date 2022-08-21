package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewManagementCmd(t *testing.T) {
	expected := &cobra.Command{
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
	assert.ObjectsAreEqual(expected, NewManagementCmd())
}
