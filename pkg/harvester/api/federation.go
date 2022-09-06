package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
)

func (a *HTTPApi) GetFederationRelationships(ctx echo.Context, params harvester.GetFederationRelationshipsParams) error {
	a.logger.Info(telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.List, params)
	telemetry.Count(ctx.Request().Context(), telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get)

	var result []common.FederationRelationship
	var fr common.FederationRelationship

	result = append(result, fr)

	return ctx.JSON(http.StatusOK, result)
}

func (a *HTTPApi) GetFederationRelationshipByID(ctx echo.Context, relationshipID int64) error {
	a.logger.Info(telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get, relationshipID)
	telemetry.Count(ctx.Request().Context(), telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Get)

	var result common.FederationRelationship

	if result == (common.FederationRelationship{}) {
		a.logger.Error("Federation relationship not found for id", relationshipID)
		return echo.NewHTTPError(http.StatusNotFound, "Federation relationship not found")
	}

	a.logger.Info("Federation relationship found for id", relationshipID)
	return ctx.JSON(http.StatusOK, result)
}

func (a *HTTPApi) UpdateFederatedRelationshipConsent(ctx echo.Context, relationshipID int64) error {
	a.logger.Info(telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Update, relationshipID)
	telemetry.Count(ctx.Request().Context(), telemetry.HTTPApi, telemetry.FederationRelationship, telemetry.Update)

	var result common.FederationRelationship

	return ctx.JSON(http.StatusOK, result)
}
