package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			runHarvesterAPI()
		},
	}
}

func runHarvesterAPI() {
	cfg, err := config.LoadFromDisk(defaultConfPath)
	if err != nil {
		fmt.Print("Error loading config:", err)
	}

	ctx := context.Background()
	harvester.NewHarvesterManager().Start(ctx, *cfg)
}

func init() {
	RootCmd.AddCommand(NewRunCmd())
}
