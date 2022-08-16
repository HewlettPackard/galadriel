package main

import (
	"os"

	"github.com/HewlettPackard/galadriel/cmd/server/cli"
)

func main() {
	os.Exit(cli.Execute())
}
