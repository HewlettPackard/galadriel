package cli

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <member | relationship>",
	Short: "Allows creation of members and relationships",
}

var createMemberCmd = &cobra.Command{
	Use:   "member <trust-domain>",
	Args:  cobra.ExactArgs(1),
	Short: "Creates a new member for the given trust domain.",

	RunE: func(cmd *cobra.Command, args []string) error {
		td := args[0]
		c := util.NewServerClient(defaultSocketPath)

		if err := c.CreateMember(&common.Member{TrustDomain: td}); err != nil {
			return err
		}

		fmt.Printf("Member created for trust domain: %s\n", td)

		return nil
	},
}

var createRelationshipCmd = &cobra.Command{
	Use:   "relationship <trust-domain-A> <trust-domain-B>",
	Short: "Registers a new relationship.",
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)

		if err := c.CreateRelationship(&common.Relationship{
			TrustDomainA: args[0],
			TrustDomainB: args[1],
		}); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	createCmd.AddCommand(createRelationshipCmd)
	createCmd.AddCommand(createMemberCmd)

	RootCmd.AddCommand(createCmd)
}
