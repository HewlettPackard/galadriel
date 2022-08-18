package main

import (
	"fmt"
	"os"

	"github.com/HewlettPackard/galadriel/cmd/harvester/cli"
)

func main() {
	fmt.Print("main")
	harvesterCLI := cli.NewHarvesterCli()
	os.Exit(harvesterCLI.Run())
}
