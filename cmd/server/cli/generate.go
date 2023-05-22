package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Args:  cobra.ExactArgs(0),
	Short: "Generates a join token for provided trust domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		trustDomain, err := cmd.Flags().GetString("trustdomain")
		if err != nil {
			return fmt.Errorf("cannot get trust domain flag: %v", err)
		}

		client, err := util.NewServerClient(defaultSocketPath)
		if err != nil {
			return err
		}

		joinToken, err := client.GetJoinToken(context.Background(), trustDomain)
		if err != nil {
			return err
		}

		fmt.Printf("Token: %s\n", joinToken.Token)
		return nil
	},
}

func init() {
	generateCmd.AddCommand(tokenCmd)
	tokenCmd.PersistentFlags().StringP("trustdomain", "t", "", "A trust domain which the join token is bound to.")
	RootCmd.AddCommand(generateCmd)
}
