package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var createCmd = &cobra.Command{
	Use:   "create <trustdomain | relationship>",
	Short: "Allows creation of trust domains and relationships",
}

var createTrustDomainCmd = &cobra.Command{
	Use:   "trustdomain",
	Args:  cobra.ExactArgs(0),
	Short: "Creates a new trust domain",

	RunE: func(cmd *cobra.Command, args []string) error {
		trustDomain, err := cmd.Flags().GetString("trustDomain")
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		client, err := util.NewServerClient(defaultSocketPath)
		if err != nil {
			return err
		}

		trustDomainRes, err := client.CreateTrustDomain(context.Background(), trustDomain)
		if err != nil {
			return err
		}

		fmt.Printf("Trust Domain created: %s\n", trustDomainRes.Name.String())

		return nil
	},
}

var createRelationshipCmd = &cobra.Command{
	Use:   "relationship",
	Short: "Registers a bidirectional federation relationship between two trust domains.",
	Args:  cobra.ExactArgs(0),

	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := util.NewServerClient(defaultSocketPath)
		if err != nil {
			return err
		}

		tdA, err := cmd.Flags().GetString("trustDomainA")
		if err != nil {
			return fmt.Errorf("cannot get trust domain A flag: %v", err)
		}

		trustDomain1, err := spiffeid.TrustDomainFromString(tdA)
		if err != nil {
			return fmt.Errorf("failed parsing trust domain: %v", err)
		}

		tdB, err := cmd.Flags().GetString("trustDomainB")
		if err != nil {
			return fmt.Errorf("cannot get trust domain B flag: %v", err)
		}

		trustDomain2, err := spiffeid.TrustDomainFromString(tdB)
		if err != nil {
			return fmt.Errorf("failed parsing trust domain: %v", err)
		}

		_, err = client.CreateRelationship(context.Background(), &entity.Relationship{
			TrustDomainAName: trustDomain1,
			TrustDomainBName: trustDomain2,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Relationship created between trust domains %s and %s\n", tdA, tdB)
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
