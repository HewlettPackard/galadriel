package cli

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
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
		trustDomain, err := spiffeid.TrustDomainFromString(td)
		if err != nil {
			return err
		}

		c := util.NewServerClient(defaultSocketPath)

		if err := c.CreateMember(&common.Member{TrustDomain: trustDomain}); err != nil {
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

		td1 := args[0]
		trustDomain1, err := spiffeid.TrustDomainFromString(td1)
		if err != nil {
			return err
		}

		td2 := args[1]
		trustDomain2, err := spiffeid.TrustDomainFromString(td2)
		if err != nil {
			return err
		}

		if err := c.CreateRelationship(&common.Relationship{
			TrustDomainA: trustDomain1,
			TrustDomainB: trustDomain2,
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
