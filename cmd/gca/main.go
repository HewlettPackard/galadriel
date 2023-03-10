package main

import (
	"os"

	"github.com/HewlettPackard/galadriel/cmd/gca/cli"
)

func main() {
	os.Exit(cli.Run())
}
