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
	Use:   "member",
	Args:  cobra.ExactArgs(0),
	Short: "Creates a new member for the given trust domain.",

	RunE: func(cmd *cobra.Command, args []string) error {
		td, err := cmd.Flags().GetString("trustDomain")
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		trustDomain, err := spiffeid.TrustDomainFromString(td)
		if err != nil {
			return err
		}

		c := util.NewServerClient(defaultSocketPath)

		if err := c.CreateMember(&common.Member{TrustDomain: trustDomain}); err != nil {
			return err
		}

		fmt.Printf("Member created for trust domain: %q\n", trustDomain.String())

		return nil
	},
}

var createRelationshipCmd = &cobra.Command{
	Use:   "relationship",
	Short: "Registers a bidirectional federation relationship between two trust domains.",
	Args:  cobra.ExactArgs(0),

	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)

		td1, err := cmd.Flags().GetString("trustDomainA")
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		trustDomain1, err := spiffeid.TrustDomainFromString(td1)
		if err != nil {
			return err
		}

		td2, err := cmd.Flags().GetString("trustDomainB")
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}
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

		fmt.Printf("Relationship created between trust domain %q and trust domain %q\n", trustDomain1.String(), trustDomain2.String())
		return nil
	},
}

func init() {
	createCmd.AddCommand(createRelationshipCmd)
	createCmd.AddCommand(createMemberCmd)

	createMemberCmd.PersistentFlags().StringP("trustDomain", "t", "", "The trust domain represented by the member.")

	createRelationshipCmd.PersistentFlags().StringP("trustDomainA", "a", "", "A trust domain to participate in a relationship.")
	createRelationshipCmd.PersistentFlags().StringP("trustDomainB", "b", "", "A trust domain to participate in a relationship.")

	RootCmd.AddCommand(createCmd)
}
