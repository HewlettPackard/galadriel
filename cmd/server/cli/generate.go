package cli

import (
	"fmt"
	"strings"

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
	Short: "Generates a join token for provided trust domain",
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

		joinToken, err := c.GenerateJoinToken(trustDomain)
		if err != nil {
			return err
		}

		fmt.Printf("Token: %s", strings.ReplaceAll(joinToken, "\"", ""))
		return nil
	},
}

func init() {
	generateCmd.AddCommand(tokenCmd)
	tokenCmd.PersistentFlags().StringP("trustDomain", "t", "", "A trust domain which the join token is bound to.")
	RootCmd.AddCommand(generateCmd)
}
