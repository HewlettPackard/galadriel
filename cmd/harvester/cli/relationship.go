package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"
	"github.com/HewlettPackard/galadriel/cmd/harvester/util"
	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var validConsentStatusValues = []string{"approved", "denied", "pending"}

var relationshipCmd = &cobra.Command{
	Use:   "relationship",
	Args:  cobra.ExactArgs(0),
	Short: "Manage relationships for the trust domain managed by the SPIRE Server this Harvester runs alongside",
	Long: `
The 'relationship' command allows you to manage relationships within the trust domain 
managed by the SPIRE Server that this Harvester runs alongside.

As the Harvester agent runs alongside the SPIRE Server, it plays a crucial role in 
managing the trust domain. The trust domain is under the management of the SPIRE Server, 
and this Harvester agent facilitates secure communication with the Galadriel Server 
to establish and manage federation relationships.

Using the 'relationship' command, you can view, approve, and deny relationships 
within the trust domain. These relationships enable secure communication 
across trust boundaries within your distributed system.

Please exercise caution when managing relationships, as they are essential for maintaining 
the security and integrity of the trust domain.
`,
}

var listRelationshipCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List relationships for the trust domain managed by the SPIRE Server this Harvester runs alongside",
	Long: `     
The 'list' command allows you to retrieve a list of relationships within the trust domain managed by the SPIRE Server and this Harvester.`,
	Example: "relationship list -s approved",
	RunE: func(cmd *cobra.Command, args []string) error {
		socketPath, err := cmd.Flags().GetString(cli.SocketPathFlagName)
		if err != nil {
			return fmt.Errorf("cannot get socket path flag: %v", err)
		}

		status, err := cmd.Flags().GetString(cli.ConsentStatusFlagName)
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		client, err := util.NewUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		consentStatus := api.ConsentStatus(status)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		relationships, err := client.GetRelationships(ctx, consentStatus)
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

var approveRelationshipCmd = &cobra.Command{
	Use:   "approve",
	Args:  cobra.ExactArgs(0),
	Short: "Approve a relationship",
	Long: `
The 'approve' command allows you to approve a pending relationship in the trust domain 
managed by the SPIRE Server and this Harvester.

By executing this command and specifying the relationship ID, you can approve a pending 
relationship between the trust domain and another registered trust domain. Once approved, 
the relationship enables secure communication across trust boundaries, 
leveraging SPIFFE Federation.

Please exercise caution when approving relationships, 
as they can have implications on the security and integrity of your distributed system. 
Ensure that you verify the authenticity and trustworthiness of the relationship before 
approving it.
`,
	Example: "relationship approve --relationshipID <relationshipID>",
	RunE: func(cmd *cobra.Command, args []string) error {
		return modifyRelationship(cmd, args, api.Approved)
	},
}

var denyRelationshipCmd = &cobra.Command{
	Use:   "deny",
	Args:  cobra.ExactArgs(0),
	Short: "Deny a relationship",
	Long: `
The 'deny' command allows you to deny a pending relationship in the trust domain 
managed by the SPIRE Server and this Harvester.

By executing this command and specifying the relationship ID, you can deny a pending 
relationship request between the trust domain and another registered trust domain. 
Denying a relationship request indicates that the trust domain does not approve 
the establishment of a federation relationship.

Please exercise caution when denying relationships, as they can have implications on the 
trust your distributed system. Ensure that you carefully evaluate the relationship request 
before denying it.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return modifyRelationship(cmd, args, api.Denied)
	},
}

func modifyRelationship(cmd *cobra.Command, args []string, action api.ConsentStatus) error {
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

	client, err := util.NewUDSClient(socketPath, nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rel, err := client.UpdateRelationship(ctx, relID, action)
	if err != nil {
		return err
	}

	switch action {
	case api.Approved:
		fmt.Print("Successfully approved relationship.\n\n")
	case api.Denied:
		fmt.Print("Successfully denied relationship.\n\n")
	}
	fmt.Printf("%s\n", rel.ConsoleString())

	return nil
}

func init() {
	RootCmd.AddCommand(relationshipCmd)
	relationshipCmd.AddCommand(listRelationshipCmd)
	relationshipCmd.AddCommand(approveRelationshipCmd)
	relationshipCmd.AddCommand(denyRelationshipCmd)

	approveRelationshipCmd.Flags().StringP(cli.RelationshipIDFlagName, "r", "", "Relationship ID to approve")
	err := approveRelationshipCmd.MarkFlagRequired(cli.RelationshipIDFlagName)
	if err != nil {
		fmt.Printf("cannot mark relationshipID flag as required: %v", err)
	}

	denyRelationshipCmd.Flags().StringP(cli.RelationshipIDFlagName, "r", "", "Relationship ID to deny")
	err = denyRelationshipCmd.MarkFlagRequired(cli.RelationshipIDFlagName)
	if err != nil {
		fmt.Printf("cannot mark relationshipID flag as required: %v", err)
	}

	listRelationshipCmd.Flags().StringP(cli.ConsentStatusFlagName, "s", "", fmt.Sprintf("Consent status to filter relationships by. Valid values: %s", strings.Join(validConsentStatusValues, ", ")))
	listRelationshipCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		status, err := cmd.Flags().GetString(cli.ConsentStatusFlagName)
		if err != nil {
			return fmt.Errorf("cannot get status flag: %v", err)
		}
		if status != "" {
			return validateConsentStatusValue(status)
		}
		return nil
	}
}

func validateConsentStatusValue(status string) error {
	for _, validValue := range validConsentStatusValues {
		if status == validValue {
			return nil
		}
	}
	return fmt.Errorf("invalid value for status. Valid values: %s", strings.Join(validConsentStatusValues, ", "))
}
