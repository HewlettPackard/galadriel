package cli

// import (
// 	"context"
// 	"fmt"

// 	"github.com/HewlettPackard/galadriel/pkg/harvester"
// 	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
// 	"github.com/spf13/cobra"
// )

// func NewRunCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "run",
// 		Short: "Runs the Galadriel server",
// 		Long:  "Run this command to start the Galadriel server",
// 	}
// }

// func init() {
// 	fmt.Printf("API")
// 	HarvesterCmd.AddCommand(NewRunCmd())
// }

// func runHarvesterAPI() {
// 	var cfg *config.HarvesterConfig
// 	ctx := context.Background()
// 	harvester.NewHarvesterManager().Start(ctx, *cfg)
// }
