package main

import (
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {

	var port = flag.Int("port", 33208, "Port for HTTP Galadriel server")
	flag.Parse()

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.

	// galadriel_server := management.MyDumbServer{}

	// This is how you set up a basic Echo router
	router := echo.New()

	// Log all requests
	router.Use(echomiddleware.Logger())

	// We now register our store above as the handler for the interface
	// management.RegisterHandlers(router, galadriel_server)

	// And we serve HTTP until the world ends.
	router.Logger.Fatal(router.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
