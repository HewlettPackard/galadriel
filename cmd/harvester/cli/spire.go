package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
	"github.com/spf13/cobra"
)

func NewSpireCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spire-list-federations",
		Short: "Manage SPIRE Server",
		Long:  "Manages the SPIRE Server the Harvester is connected to",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: arguments reading and config parsing should be done somewhere else
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("error reading config file path: %v", err)
			}
			cfg, err := config.LoadFromDisk(configPath)
			if err != nil {
				return err
			}

			ctx := context.Background()

			spire := spire.NewLocalSpireServer(ctx, cfg.HarvesterConfigSection.SpireSocketPath)
			feds, err := spire.ListFederationRelationships(ctx)
			if err != nil {
				return fmt.Errorf("error getting federation relationships: %v", err)
			}
			if len(feds) == 0 {
				return errors.New("no federation relationships found")
			}

			for _, fed := range feds {
				fmt.Printf("Trust Domain: %s\n", fed.TrustDomain)
				fmt.Printf("Bundle Endpoint Profile: %T\n", fed.BundleEndpointProfile)
				fmt.Printf("Bundle Endpoint URL: %s\n", fed.BundleEndpointURL)
				fmt.Println("------")
			}

			return nil
		},
	}
}

func init() {
	runCmd := NewSpireCmd()
	runCmd.PersistentFlags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file")
	RootCmd.AddCommand(runCmd)
}
