package cli

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var createCmd = &cobra.Command{
	Use:   "create <trustdomain| relationship>",
	Short: "Allows creation of trust domains and relationships",
}

var createTrustDomainCmd = &cobra.Command{
	Use:   "trustdomain",
	Args:  cobra.ExactArgs(0),
	Short: "Creates a new trust domain for the given trust domain.",

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

		if err := c.CreateTrustDomain(&entity.TrustDomain{Name: trustDomain}); err != nil {
			return err
		}

		fmt.Printf("Trust Domain created for trust domain: %q\n", trustDomain.String())

		return nil
	},
}

var createRelationshipCmd = &cobra.Command{
	Use:   "relationship",
	Short: "Registers a bidirectional federation relationship between two trust domains.",
	Args:  cobra.ExactArgs(0),

	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)

		tdA, err := cmd.Flags().GetString("trustDomainA")
		if err != nil {
			return fmt.Errorf("cannot get trust domain A flag: %v", err)
		}

		trustDomain1, err := spiffeid.TrustDomainFromString(tdA)
		if err != nil {
			return err
		}

		tdb, err := cmd.Flags().GetString("trustDomainB")
		if err != nil {
			return fmt.Errorf("cannot get trust domain B flag: %v", err)
		}
		trustDomain2, err := spiffeid.TrustDomainFromString(tdb)
		if err != nil {
			return err
		}

		if err := c.CreateRelationship(&entity.Relationship{
			TrustDomainAName: trustDomain1,
			TrustDomainBName: trustDomain2,
		}); err != nil {
			return err
		}

		fmt.Printf("Relationship created between trust domain %q and trust domain %q\n", trustDomain1.String(), trustDomain2.String())
		return nil
	},
}

func init() {
	createCmd.AddCommand(createRelationshipCmd)
	createCmd.AddCommand(createTrustDomainCmd)

	createTrustDomainCmd.PersistentFlags().StringP("trustDomain", "t", "", "The trust domain name.")

	createRelationshipCmd.PersistentFlags().StringP("trustDomainA", "a", "", "A trust domain name to participate in a relationship.")
	createRelationshipCmd.PersistentFlags().StringP("trustDomainB", "b", "", "A trust domain name to participate in a relationship.")

	RootCmd.AddCommand(createCmd)
}
