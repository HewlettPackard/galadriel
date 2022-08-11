package controller

import (
	"context"
	"errors"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
)

type HarvesterController interface {
	common.RunnablePlugin
	GetTrustBundle(context.Context, string) (string, error)
}

type LocalHarvesterController struct {
	logger  common.Logger
	catalog catalog.Catalog
}

func NewLocalHarvesterController(catalog catalog.Catalog) HarvesterController {
	return &LocalHarvesterController{
		logger:  *common.NewLogger("harvester_controller"),
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
	return "", errors.New("not implemented")
}

func (c *LocalHarvesterController) run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-ticker.C:
			c.logger.Debug("Doing something")
		case <-ctx.Done():
			c.logger.Debug("Done")
			return
		}
	}
}
