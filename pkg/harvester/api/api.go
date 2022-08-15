package api

import (
	"context"
	"errors"

	"github.com/HewlettPackard/Galadriel/pkg/common"
	"github.com/HewlettPackard/Galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/Galadriel/pkg/harvester/controller"
)

type API interface {
	common.RunnablePlugin
	GetTrustBundle(context.Context, string) (string, error)
	AddTrustBundle(context.Context, string) error
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

func (a *HTTPApi) GetTrustBundle(ctx context.Context, spiffeID string) (string, error) {
	return "", errors.New("not implemented")
}

func (a *HTTPApi) AddTrustBundle(ctx context.Context, spiffeID string) error {
	return errors.New("not implemented")
}
