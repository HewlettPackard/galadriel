package cli

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
}

var tokenCmd = &cobra.Command{
	Use:   "token <memberID>",
	Args:  cobra.ExactArgs(1),
	Short: "Generates an access token for provided memberID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := util.NewServerClient(defaultSocketPath)
		memberID, err := uuid.Parse(args[0])
		if err != nil {
			return err
		}
		at, err := c.GenerateAccessToken(memberID)
		if err != nil {
			return err
		}

		fmt.Println(at.Token)
		return nil
	},
}

func init() {
	generateCmd.AddCommand(tokenCmd)
	RootCmd.AddCommand(generateCmd)
}
