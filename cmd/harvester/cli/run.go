package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the Galadriel server",
		Long:  "Run this command to start the Galadriel server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("sou o run")
			// runHarvesterAPI()
		},
	}
}

func init() {
	fmt.Printf("API")
	HarvesterCmd.AddCommand(NewRunCmd())
}
