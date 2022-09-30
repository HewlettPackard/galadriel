package cli

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var generateCmd = &cobra.Command{
	Use: "generate",
}

var tokenCmd = &cobra.Command{
	Use:   "token <trust-domain>",
	Args:  cobra.ExactArgs(1),
	Short: "Generates an access token for provided trust domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)

		td := args[0]
		trustDomain, err := spiffeid.TrustDomainFromString(td)
		if err != nil {
			return err
		}

		at, err := c.GenerateAccessToken(trustDomain)
		if err != nil {
			return err
		}

		fmt.Println("Access Token: " + at.Token)
		return nil
	},
}

func init() {
	generateCmd.AddCommand(tokenCmd)
	RootCmd.AddCommand(generateCmd)
}
