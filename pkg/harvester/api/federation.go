package api

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
)

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
