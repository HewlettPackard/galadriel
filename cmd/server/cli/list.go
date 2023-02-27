package cli

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <trustDomains | relationships>",
	Short: "Lists trustDomains and relationships",
}

var listTrustDomainCmd = &cobra.Command{
	Use:   "trustdomains",
	Args:  cobra.ExactArgs(0),
	Short: "Lists all the Trust Domains.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)
		trustDomains, err := c.ListTrustDomains()
		if err != nil {
			return err
		}

		if len(trustDomains) == 0 {
			fmt.Println("No trust domains found")
			return nil
		}

		for _, m := range trustDomains {
			fmt.Printf("ID: %s\n", m.ID.UUID)
			fmt.Printf("Trust Domain: %s\n", m.Name)
			fmt.Println()
		}

		return nil
	},
}

var listRelationshipsCmd = &cobra.Command{
	Use:   "relationships",
	Args:  cobra.ExactArgs(0),
	Short: "Lists all the relationships.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)
		rels, err := c.ListRelationships()
		if err != nil {
			return err
		}

		if len(rels) == 0 {
			fmt.Println("No relationships found")
			return nil
		}

		for _, r := range rels {
			fmt.Printf("ID: %s\n", r.ID.UUID)
			fmt.Printf("Trust Domain A: %s\n", r.TrustDomainAName.String())
			fmt.Printf("Trust Domain B: %s\n", r.TrustDomainBName.String())
			fmt.Println()
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listTrustDomainCmd)
	listCmd.AddCommand(listRelationshipsCmd)

	RootCmd.AddCommand(listCmd)
}
