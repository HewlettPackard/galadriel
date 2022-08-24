// package main

// import (
// 	"os"

// 	"github.com/n-th/galadriel/cmd/server/cli"
// )

// func main() {
// 	os.Exit(new(cli.CLI).Run(os.Args[1:]))
// }

package main

import (
	"flag"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/server/api/management"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {

	var port = flag.Int("port", 33208, "Port for HTTP Galadriel server")
	flag.Parse()

	// This is how you set up a basic Echo router
	router := echo.New()

	// Create a server instance
	galadrielServer := management.NewManagement()

	// Register router as handler for API interface and routes
	management.RegisterHandlers(router, galadrielServer)

	// Log all requests
	router.Use(echomiddleware.Logger())

	// And we serve HTTP until the world ends.
	router.Logger.Fatal(router.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
