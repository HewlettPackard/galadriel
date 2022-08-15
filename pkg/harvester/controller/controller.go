package controller

import (
	"context"
	"errors"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
)

type HarvesterController interface {
	common.RunnablePlugin
	GetTrustBundle(context.Context, string) (string, error)
	AddTrustBundle(context.Context, string) error
}

type LocalHarvesterController struct {
	logger  common.Logger
	catalog catalog.Catalog
}

func NewLocalHarvesterController(catalog catalog.Catalog) HarvesterController {
	return &LocalHarvesterController{
		logger:  *common.NewLogger(telemetry.HarvesterController),
		catalog: catalog,
	}
}

func (c *LocalHarvesterController) Run(ctx context.Context) error {
	c.logger.Info("Starting harvester controller")

	go c.run(ctx)

	<-ctx.Done()
	return nil
}

func (c *LocalHarvesterController) GetTrustBundle(ctx context.Context, spiffeID string) (string, error) {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Get)
	return "", errors.New("not implemented")
}

func (c *LocalHarvesterController) AddTrustBundle(ctx context.Context, spiffeID string) error {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Add)
	return errors.New("not implemented")
}

func (c *LocalHarvesterController) run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-ticker.C:
			c.logger.Debug("Doing something")
			telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Add)
		case <-ctx.Done():
			c.logger.Debug("Done")
			return
		}
	}
}
