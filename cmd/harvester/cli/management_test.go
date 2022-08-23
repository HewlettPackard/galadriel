package cli

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewManagementCmd(t *testing.T) {
	expected := &cobra.Command{
		Use:   "manage",
		Short: "Manage member and federation relationships",
		Long:  "Run this command to approve and deny relationships between members or federations",
		RunE: func(cmd *cobra.Command, args []string) error {
			HarvesterCLI.runManagementAPI()
			return errors.New("not implemented")
		},
	}
	assert.ObjectsAreEqual(expected, NewManagementCmd())
}
