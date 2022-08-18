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
	Short: "Commands to run Galadriel Harvester",
	Long:  "Commands to manage Galadriel Harvester",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("sou o run")
		runHarvesterAPI()
	},
}

func init() {
	fmt.Print("server")
	harvesterCli := NewHarvesterCli()
	harvesterCli.AddCommand(HarvesterCmd)
}

func runHarvesterAPI() {
	var cfg *config.HarvesterConfig
	ctx := context.Background()
	harvester.NewHarvesterManager().Start(ctx, *cfg)
}
