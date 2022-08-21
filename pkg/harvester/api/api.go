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

func (a *HTTPApi) GetTrustBundle(ctx context.Context, spiffeID string) (common.TrustBundle, error) {
	telemetry.Count(ctx, telemetry.HTTPApi, telemetry.TrustBundle, telemetry.Get)

	var tb common.TrustBundle

	tb, err := a.controller.GetTrustBundle(ctx, spiffeID)
	if err != nil {
		a.logger.Error(err)
		return tb, err
	}

	return tb, nil
}

func (a *HTTPApi) AddTrustBundle(ctx context.Context, trustBundle common.TrustBundle) (common.TrustBundle, error) {
	telemetry.Count(ctx, telemetry.HTTPApi, telemetry.TrustBundle, telemetry.Add)

	tb, err := a.controller.AddTrustBundle(ctx, trustBundle)
	if err != nil {
		a.logger.Error(err)
		return tb, err
	}

	return tb, nil
}

// POST: federation-relationship/{relationshipId} {action: approve/deny}
func (a *HTTPApi) ManageFederationRelationship(ctx context.Context, id string) (common.FederationRelationship, error) {
	var fr common.FederationRelationship

	// if body.action == telemetry.Approve {
	// 	telemetry.Count(ctx, telemetry.HTTPApi, telemetry.TrustBundle, telemetry.Approve)
	// 	fr, err := a.controller.ApproveFederationRelationship(ctx, id)
	//  if err != nil {
	//    a.logger.Error(err)
	//    return fr, err
	//   }
	// }

	// if body.action == telemetry.Deny {
	// 	telemetry.Count(ctx, telemetry.HTTPApi, telemetry.TrustBundle, telemetry.Approve)
	//  fr, err := a.controller.DenyFederationRelationship(ctx, id)
	//  if err != nil {
	//    a.logger.Error(err)
	//    return fr, err
	//  }
	// }

	return fr, nil
}

func (a *HTTPApi) GetFederationRelationshipsByStatus(ctx context.Context, status string) ([]common.FederationRelationship, error) {
	telemetry.Count(ctx, telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get)

	var fr []common.FederationRelationship

	fr, err := a.controller.GetFederationRelationshipsByStatus(ctx, status)
	if err != nil {
		a.logger.Error(err)
		return fr, err
	}

	return fr, nil
}
