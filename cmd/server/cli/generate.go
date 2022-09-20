package cli

import (
	"fmt"
	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Generates a join token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)
		token, err := c.GenerateJoinToken()
		if err != nil {
			return err
		}

		fmt.Println(token)
		return nil
	},
}

func init() {
	generateCmd.AddCommand(tokenCmd)
	RootCmd.AddCommand(generateCmd)
}
