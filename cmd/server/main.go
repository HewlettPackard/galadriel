package main

import (
	"os"

	"github.com/HewlettPackard/Galadriel/cmd/server/cli"
)

func main() {
	os.Exit(cli.Execute())
}
