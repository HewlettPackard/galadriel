package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"
	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/spf13/cobra"
)

const trustDomainCommonText = `
A trust domain represents a distinct trust boundary or realm within a distributed system. 
By creating a new trust domain, you can establish a namespace for workload identities 
and define the security boundaries for your system.

Registering a trust domain in Galadriel Server enables secure interactions 
with other registered trust domains through SPIFFE Federation relationships. It allows 
for the exchange of Federated bundles and facilitates seamless communication between 
trust domains.
`

var trustDomainCmd = &cobra.Command{
	Use:   "trustdomain",
	Short: "Manage SPIFFE trust domains",
	Long: `
The 'trustdomain' command is used for managing SPIFFE trust domains in the Galadriel 
Server database. It allows you to register, list, update, and delete trust domains.
` + trustDomainCommonText + `
`,
}

var createTrustDomainCmd = &cobra.Command{
	Use:   "create",
	Args:  cobra.ExactArgs(0),
	Short: "Create a new trust domain",
	Long: `
The 'create' command allows you to create a new trust domain in the Galadriel Server.
` + trustDomainCommonText + `
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		socketPath, err := cmd.Flags().GetString(cli.SocketPathFlagName)
		if err != nil {
			return fmt.Errorf("cannot get socket path flag: %v", err)
		}

		trustDomain, err := cmd.Flags().GetString(cli.TrustDomainFlagName)
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		if trustDomain == "" {
			return fmt.Errorf("trust domain name is required")
		}

		client, err := util.NewGaladrielUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		trustDomainRes, err := client.CreateTrustDomain(ctx, trustDomain)
		if err != nil {
			return err
		}

		fmt.Printf("Trust Domain created: %s\n", trustDomainRes.Name.String())

		return nil
	},
}

var listTrustDomainCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List trust domains",
	Long:  `The 'list' command allows you to retrieve a list of registered trust domains.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var deleteTrustDomainCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(0),
	Short: "Delete a trust domain",
	Long: `The 'delete' command allows you to remove a trust domain from the Galadriel Server.

Before deleting a trust domain, ensure that all federation relationships associated 
with it are removed or deleted. This ensures the integrity of the system and prevents 
potential disruptions in secure communication between trust domains.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var updateTrustDomainCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.ExactArgs(0),
	Short: "Update a trust domain",
	Long: `The 'update' command allows you to modify the configuration of a trust domain 
in the Galadriel Server.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(trustDomainCmd)
	trustDomainCmd.AddCommand(createTrustDomainCmd)
	trustDomainCmd.AddCommand(listTrustDomainCmd)
	trustDomainCmd.AddCommand(deleteTrustDomainCmd)
	trustDomainCmd.AddCommand(updateTrustDomainCmd)

	createTrustDomainCmd.Flags().StringP(cli.TrustDomainFlagName, "t", "", "The trust domain name.")
	err := createTrustDomainCmd.MarkFlagRequired(cli.TrustDomainFlagName)
	if err != nil {
		fmt.Printf("Error marking trustDomain flag as required: %v\n", err)
	}
}
