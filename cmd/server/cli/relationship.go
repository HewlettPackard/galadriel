package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"
	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	relationshipCommonText = `Manage federation relationships between SPIFFE trust domains with the 'relationship' command.
Federation relationships in SPIFFE permit secure communication between workloads across different trust domains.`
)

var relationshipCmd = &cobra.Command{
	Use:   "relationship",
	Short: "Manage federation relationships of trust domains",
	Long: `
` + relationshipCommonText + `
`,
}

var createRelationshipCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new federation relationship",
	Long: ` 
The 'create' command initiates a new federation relationship between two specified SPIFFE trust domains.

Importantly, the initiation of a federation relationship is a two-party agreement: it needs to be approved by both trust domains involved.

` + relationshipCommonText + `
`,
	Args: cobra.ExactArgs(0),

	RunE: func(cmd *cobra.Command, args []string) error {
		socketPath, err := cmd.Flags().GetString(cli.SocketPathFlagName)
		if err != nil {
			return fmt.Errorf("cannot get socket path flag: %v", err)
		}

		client, err := util.NewGaladrielUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		tdA, err := cmd.Flags().GetString(cli.TrustDomainAFlagName)
		if err != nil {
			return fmt.Errorf("cannot get trust domain A flag: %v", err)
		}

		if tdA == "" {
			return fmt.Errorf("trust domain A flag is required")
		}

		trustDomain1, err := spiffeid.TrustDomainFromString(tdA)
		if err != nil {
			return err
		}

		tdB, err := cmd.Flags().GetString(cli.TrustDomainBFlagName)
		if err != nil {
			return fmt.Errorf("cannot get trust domain B flag: %v", err)
		}

		if tdB == "" {
			return fmt.Errorf("trust domain B flag is required")
		}

		trustDomain2, err := spiffeid.TrustDomainFromString(tdB)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = client.CreateRelationship(ctx, &entity.Relationship{
			TrustDomainAName: trustDomain1,
			TrustDomainBName: trustDomain2,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Relationship created between trust domains %q and %q\n", tdA, tdB)
		return nil
	},
}

var listRelationshipCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List relationships",
	Long:  `The 'list' command allows you to retrieve a list of registered relationships.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var deleteRelationshipCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(0),
	Short: "Delete a relationship",
	Long: `
The 'delete' command allows you to remove a relationship from the Galadriel Server.

By specifying the relationship to delete, this command removes the corresponding relationship configuration from the Galadriel Server. However, please note that deleting a relationship can lead to potential disruptions in secure communication between trust domains.

Before deleting a relationship, carefully consider the implications it may have on the trust and connectivity between the associated trust domains. Ensure that the removal of the relationship aligns with your system's security requirements and communication needs.

Exercise caution when using this command, as it permanently removes the relationship configuration and may affect the ability of workloads in different trust domains to securely communicate with each other.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var updateRelationshipCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.ExactArgs(0),
	Short: "Update a trust domain",
	Long: `The 'update' command allows you to modify the configuration of a relationship
in the Galadriel Server.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(relationshipCmd)
	relationshipCmd.AddCommand(createRelationshipCmd)
	relationshipCmd.AddCommand(listRelationshipCmd)
	relationshipCmd.AddCommand(deleteRelationshipCmd)
	relationshipCmd.AddCommand(updateRelationshipCmd)

	createRelationshipCmd.Flags().StringP(cli.TrustDomainAFlagName, "a", "", "The name of a SPIFFE trust domain to participate in the relationship.")
	createRelationshipCmd.Flags().StringP(cli.TrustDomainBFlagName, "b", "", "The name of a SPIFFE trust domain to participate in the relationship.")
}
