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
	Use:   "token",
	Args:  cobra.ExactArgs(0),
	Short: "Generates an access token for provided trust domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)

		td, err := cmd.Flags().GetString("trustDomain")
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

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
	tokenCmd.PersistentFlags().StringP("trustDomain", "t", "", "A trust domain which the access token is bound to.")
	RootCmd.AddCommand(generateCmd)
}
