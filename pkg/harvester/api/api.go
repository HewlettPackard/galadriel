package api

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
)

type API interface {
	common.RunnablePlugin
	GetTrustBundle(context.Context, string) (common.TrustBundle, error)
	AddTrustBundle(context.Context, common.TrustBundle) (common.TrustBundle, error)
	ManageFederationRelationship(context.Context, string) (common.FederationRelationship, error)
	GetFederationRelationshipsByStatus(context.Context, string) ([]common.FederationRelationship, error)
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
	// TODO: implement

	<-ctx.Done()
	return nil
}
