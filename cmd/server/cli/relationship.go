package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"
	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var (
	errMarkFlagAsRequired = "cannot mark %q flag as required: %v"
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

		trustDomain1, err := spiffeid.TrustDomainFromString(tdA)
		if err != nil {
			return err
		}

		tdB, err := cmd.Flags().GetString(cli.TrustDomainBFlagName)
		if err != nil {
			return fmt.Errorf("cannot get trust domain B flag: %v", err)
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
		socketPath, err := cmd.Flags().GetString(cli.SocketPathFlagName)
		if err != nil {
			return fmt.Errorf("cannot get socket path flag: %v", err)
		}

		status, err := cmd.Flags().GetString(cli.ConsentStatusFlagName)
		if err != nil {
			return fmt.Errorf("cannot get consent status flag: %v", err)
		}

		trustDomainName, err := cmd.Flags().GetString(cli.TrustDomainFlagName)
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		consentStatus := api.ConsentStatus(status)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, err := util.NewGaladrielUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		relationships, err := client.GetRelationships(ctx, consentStatus, trustDomainName)
		if err != nil {
			return err
		}

		if len(relationships) == 0 {
			fmt.Println("No relationships found")
			return nil
		}

		fmt.Println()
		for _, r := range relationships {
			fmt.Printf("%s\n", r.ConsoleString())
		}
		fmt.Println()

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
		socketPath, err := cmd.Flags().GetString(cli.SocketPathFlagName)
		if err != nil {
			return fmt.Errorf("cannot get socket path flag: %v", err)
		}

		idStr, err := cmd.Flags().GetString(cli.RelationshipIDFlagName)
		if err != nil {
			return fmt.Errorf("cannot get relationship ID flag: %v", err)
		}

		relID, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("cannot parse relationship ID: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, err := util.NewGaladrielUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		err = client.DeleteRelationshipByID(ctx, relID)
		if err != nil {
			return err
		}

		fmt.Printf("Relationship deleted.\n")

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
		socketPath, err := cmd.Flags().GetString(cli.SocketPathFlagName)
		if err != nil {
			return fmt.Errorf("cannot get socket path flag: %v", err)
		}

		idStr, err := cmd.Flags().GetString(cli.RelationshipIDFlagName)
		if err != nil {
			return fmt.Errorf("cannot get relationship ID flag: %v", err)
		}

		relID, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("cannot parse relationship ID: %v", err)
		}

		statusA, err := cmd.Flags().GetString(cli.ConsentStatusAFlagName)
		if err != nil {
			return fmt.Errorf("cannot get consent status for trust domain A flag: %v", err)
		}

		statusB, err := cmd.Flags().GetString(cli.ConsentStatusBFlagName)
		if err != nil {
			return fmt.Errorf("cannot get consent status for trust domain B flag: %v", err)
		}

		consentStatusA := api.ConsentStatus(statusA)
		consentStatusB := api.ConsentStatus(statusB)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client, err := util.NewGaladrielUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		rel, err := client.UpdateRelationshipByID(ctx, relID, consentStatusA, consentStatusB)
		if err != nil {
			return err
		}

		fmt.Printf("Relationship %q updated.\n", rel.ID.UUID.String())

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
	err := createRelationshipCmd.MarkFlagRequired(cli.TrustDomainAFlagName)
	if err != nil {
		fmt.Printf(errMarkFlagAsRequired, cli.TrustDomainAFlagName, err)
	}
	createRelationshipCmd.Flags().StringP(cli.TrustDomainBFlagName, "b", "", "The name of a SPIFFE trust domain to participate in the relationship.")
	err = createRelationshipCmd.MarkFlagRequired(cli.TrustDomainBFlagName)
	if err != nil {
		fmt.Printf(errMarkFlagAsRequired, cli.TrustDomainBFlagName, err)
	}

	listRelationshipCmd.Flags().StringP(cli.TrustDomainFlagName, "t", "", "The name of a trust domain to filter relationships by.")
	listRelationshipCmd.Flags().StringP(cli.ConsentStatusFlagName, "s", "", fmt.Sprintf("Consent status to filter relationships by. Valid values: %s", strings.Join(cli.ValidConsentStatusValues, ", ")))
	listRelationshipCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		status, err := cmd.Flags().GetString(cli.ConsentStatusFlagName)
		if err != nil {
			return fmt.Errorf("cannot get status flag: %v", err)
		}

		if status != "" {
			return cli.ValidateConsentStatusValue(status)
		}
		return nil
	}

	updateRelationshipCmd.Flags().StringP(cli.RelationshipIDFlagName, "r", "", "The ID of therelationship to be updated.")
	updateRelationshipCmd.Flags().StringP(cli.ConsentStatusAFlagName, "a", "", fmt.Sprintf("Trust domain A consent status to update. Valid values: %s", strings.Join(cli.ValidConsentStatusValues, ", ")))
	updateRelationshipCmd.Flags().StringP(cli.ConsentStatusBFlagName, "b", "", fmt.Sprintf("Trust domain B consent status to update. Valid values: %s", strings.Join(cli.ValidConsentStatusValues, ", ")))
	updateRelationshipCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		statusA, err := cmd.Flags().GetString(cli.ConsentStatusAFlagName)
		if err != nil {
			return fmt.Errorf("cannot get consent status A flag: %v", err)
		}
		if statusA != "" {
			return cli.ValidateConsentStatusValue(statusA)
		}

		statusB, err := cmd.Flags().GetString(cli.ConsentStatusBFlagName)
		if err != nil {
			return fmt.Errorf("cannot get consent status B flag: %v", err)
		}
		if statusB != "" {
			return cli.ValidateConsentStatusValue(statusB)
		}

		return nil
	}

	deleteRelationshipCmd.Flags().StringP(cli.RelationshipIDFlagName, "r", "", "The ID of the relationship to be deleted.")
	err = deleteRelationshipCmd.MarkFlagRequired(cli.RelationshipIDFlagName)
	if err != nil {
		fmt.Printf(errMarkFlagAsRequired, cli.RelationshipIDFlagName, err)
	}
}
