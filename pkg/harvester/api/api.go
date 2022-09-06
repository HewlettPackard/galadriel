package api

import (
	"context"
	"flag"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/labstack/echo/v4"
)

type API interface {
	common.RunnablePlugin
	GetFederationRelationships(ctx echo.Context, params harvester.GetFederationRelationshipsParams) error
	GetFederationRelationshipbyID(ctx echo.Context, relationshipID int64) error
	UpdateFederatedRelationshipConsent(ctx echo.Context, relationshipID int64) error
}

type HTTPApi struct {
	controller controller.HarvesterController
	logger     common.Logger
}

func NewHTTPApi(controller controller.HarvesterController) API {
	// TODO: Add listen address and port
	return &HTTPApi{
		controller: controller,
		logger:     *common.NewLogger(telemetry.HTTPApi),
	}
}

func (a *HTTPApi) Run(ctx context.Context) error {
	a.logger.Info("Starting HTTP API")

	var controller controller.HarvesterController

	harvester_api := NewHTTPApi(controller)

	err := a.run(ctx, harvester_api)
	if err != nil {
		a.logger.Error("HTTP API has failed", err)
		panic(err)
	}

	return nil
}

func (a *HTTPApi) run(ctx context.Context, api API) error {
	a.logger.Info("Starting HTTP API")

	var port = flag.Int("port", 8000, "Port for HTTP Galadriel server")
	flag.Parse()

	router := echo.New()
	harvester.RegisterHandlers(router, api)

	return router.Start(fmt.Sprintf("0.0.0.0:%d", *port))
}
