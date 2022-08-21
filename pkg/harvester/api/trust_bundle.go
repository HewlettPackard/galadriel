package api

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
)

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
