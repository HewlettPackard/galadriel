// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml
//
// The code under api/petstore/ has been generated from that specification.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/echo/api"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	var port = flag.Int("port", 8080, "Port for HTTP Galadriel server")
	flag.Parse()

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	store := api.NewPetStore()

	// This is how you set up a basic Echo router
	router := echo.New()

	// Log all requests
	router.Use(echomiddleware.Logger())

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	router.Use(middleware.OapiRequestValidator(swagger))

	// We now register our store above as the handler for the interface
	api.RegisterHandlers(router, store)

	// And we serve HTTP until the world ends.
	router.Logger.Fatal(router.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
