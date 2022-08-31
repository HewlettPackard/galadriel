package api

import (
	"context"
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func NewHTTPServer() HTTPServer {
	return HTTPServer{}
}

type HTTPServer struct{}

func (s HTTPServer) Run(ctx context.Context) error {
	var port = flag.Int("port", 33208, "Port for HTTP Galadriel server")
	flag.Parse()

	router := echo.New()

	// Log all requests
	router.Use(echomiddleware.Logger())

	errch := make(chan error)

	// Start serving
	go func() {
		errch <- router.Start(fmt.Sprintf("0.0.0.0:%d", *port))
	}()

	// Graceful shutdown
	var err error
	select {
	case <-ctx.Done():
		// shutdown routines
		fmt.Println("Gracefully shutting down...")
		// TODO: understand why we're not seeing logs from `router.Logger`
		router.Logger.Info("Gracefully shutting down...")
		err = router.Shutdown(ctx)
	case err = <-errch:
	}

	if err != nil {
		router.Logger.Error("Error gracefully shutting down server")
	}

	return err
}
