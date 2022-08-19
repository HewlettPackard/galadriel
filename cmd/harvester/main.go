package main

import (
	"os"

	"github.com/HewlettPackard/galadriel/cmd/harvester/cli"
)

func main() {
	HarvesterCli := cli.NewHarvesterCli()
	os.Exit(HarvesterCli.Run())
}
