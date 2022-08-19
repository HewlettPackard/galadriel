package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
	"github.com/spf13/cobra"
)

var HarvesterCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Galadriel Harvester",
	Long:  "Command to run the Galadriel Harvester API",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("sou o run")
		RunHarvesterAPI()
	},
}

func init() {
	harvesterCli := NewHarvesterCli()
	harvesterCli.AddCommand(HarvesterCmd)
}

func RunHarvesterAPI() {
	var cfg *config.HarvesterConfig
	// cfg =
	ctx := context.Background()
	harvester.NewHarvesterManager().Start(ctx, *cfg)
	// harvester.NewHarvesterManager().Start(ctx)
}
