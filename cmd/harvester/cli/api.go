package cli

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
)

const defaultConfPath = "conf/harvester/harvester.conf"

func RunHarvesterAPI() {
	cfg, err := config.LoadFromDisk(defaultConfPath)
	if err != nil {
		fmt.Print("Error loading config:", err)
	}

	ctx := context.Background()
	harvester.NewHarvesterManager().Start(ctx, *cfg)
}
