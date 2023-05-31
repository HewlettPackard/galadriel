package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"

	"github.com/HewlettPackard/galadriel/cmd/server/util"
	"github.com/HewlettPackard/galadriel/pkg/server/endpoints"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use: "token",
}

var generateTokenCmd = &cobra.Command{
	Use:   "generate",
	Args:  cobra.ExactArgs(0),
	Short: "Generates a join token for the provided trust domain",
	Long: `
The 'generate' command allows you to generate a join token for the provided trust domain.

By executing this command, you can generate a token that enables the connection of a Harvester 
to the Galadriel Server. This join token serves as a secure authentication mechanism to establish 
the necessary trust relationship between the Harvester and the Galadriel Server.

Ensure that you provide the appropriate trust domain while generating the join token to 
ensure compatibility and successful connection.

Please exercise caution when handling and sharing join tokens, as they grant access to the 
trust domain and should only be shared with authorized individuals or entities.
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

		ttlStr, err := cmd.Flags().GetString(cli.TTLFlagName)
		if err != nil {
			return fmt.Errorf("cannot get TTL flag: %v", err)
		}

		ttl, err := strconv.ParseInt(ttlStr, 10, 32)
		if err != nil {
			return errors.New("invalid TTL")
		}

		client, err := util.NewGaladrielUDSClient(socketPath, nil)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		joinToken, err := client.GetJoinToken(ctx, trustDomain, int32(ttl))
		if err != nil {
			return err
		}

		fmt.Print(joinToken.ConsoleString())
		return nil
	},
}

func init() {
	RootCmd.AddCommand(tokenCmd)

	tokenCmd.AddCommand(generateTokenCmd)

	generateTokenCmd.Flags().StringP(cli.TrustDomainFlagName, "t", "", "The trust domain to which the join token will be bound")
	err := generateTokenCmd.MarkFlagRequired(cli.TrustDomainFlagName)
	if err != nil {
		fmt.Printf("Error marking trustDomain flag as required: %v\n", err)
	}
	generateTokenCmd.Flags().StringP(cli.TTLFlagName, "", fmt.Sprintf("%d", endpoints.DefaultTokenTTL), "Token TTL in seconds")
}
