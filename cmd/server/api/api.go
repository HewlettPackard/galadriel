package api

import (
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func Run() {

	var port = flag.Int("port", 33208, "Port for HTTP Galadriel server")
	flag.Parse()

	router := echo.New()

	// Log all requests
	router.Use(echomiddleware.Logger())

	router.Logger.Fatal(router.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
