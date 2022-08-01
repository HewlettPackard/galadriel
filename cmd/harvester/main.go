package main

import (
	"os"

	"github.com/HewlettPackard/Galadriel/cmd/harvester/cli"
)

func main() {
	os.Exit(cli.NewHarvesterCLI().Run(os.Args[1:]))
}
