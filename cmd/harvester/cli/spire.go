package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
	"github.com/spf13/cobra"
)

func NewSpireCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spire-federation-list",
		Short: "List SPIRE Federation relationships",
		Long:  "List SPIRE Federation relationships",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "/tmp/spire-server/private/api.sock"
			ctx := context.Background()
			spire := spire.NewLocalSpireServer(ctx, path)
			feds, err := spire.ListFederationRelationships(ctx)
			if err != nil {
				fmt.Println("Error:", err)
			}

			fmt.Println("Feds count:", len(feds))

			return nil
		},
	}
}

func init() {
	runCmd := NewSpireCmd()
	RootCmd.AddCommand(runCmd)
}
