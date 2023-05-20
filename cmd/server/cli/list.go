package cli

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <trustdomains | relationships>",
	Short: "Lists trust domains and relationships",
}

// TODO: Implement Get Relationships and Trust Domains
var listTrustDomainCmd = &cobra.Command{}

var listRelationshipsCmd = &cobra.Command{}

func init() {
	listCmd.AddCommand(listTrustDomainCmd)
	listCmd.AddCommand(listRelationshipsCmd)

	RootCmd.AddCommand(listCmd)
}
