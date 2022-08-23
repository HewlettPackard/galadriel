package api

import (
	"github.com/labstack/echo/v4"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
)

func (a *HTTPApi) GetFederationRelationships(ctx echo.Context, params harvester.GetFederationRelationshipsParams) error {
	a.logger.Info(telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.List)
	telemetry.Count(ctx.Request().Context(), telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get)

	return nil
}

func (a *HTTPApi) GetFederationRelationshipbyId(ctx echo.Context, relationshipID int64) error {
	a.logger.Info(telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get)
	telemetry.Count(ctx.Request().Context(), telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get)

	return nil
}

func (a *HTTPApi) UpdateFederatedRelationshipConsent(ctx echo.Context, relationshipID int64) error {
	a.logger.Info(telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Update)
	telemetry.Count(ctx.Request().Context(), telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Update)

	return nil
}
