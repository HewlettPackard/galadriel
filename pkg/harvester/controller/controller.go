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
	GetTrustBundle(context.Context, string) (common.TrustBundle, error)
	AddTrustBundle(context.Context, common.TrustBundle) (common.TrustBundle, error)
	ApproveFederationRelationship(context.Context, string) (common.FederationRelationship, error)
	DenyFederationRelationship(context.Context, string) (common.FederationRelationship, error)
	GetFederationRelationshipsByStatus(context.Context, string) ([]common.FederationRelationship, error)
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

func (c *LocalHarvesterController) GetTrustBundle(ctx context.Context, spiffeID string) (common.TrustBundle, error) {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Get)
	var tb common.TrustBundle

	return tb, errors.New("not implemented")
}

func (c *LocalHarvesterController) AddTrustBundle(ctx context.Context, tb common.TrustBundle) (common.TrustBundle, error) {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Add)

	return tb, errors.New("not implemented")
}

func (c *LocalHarvesterController) ApproveFederationRelationship(ctx context.Context, spiffeID string) (common.FederationRelationship, error) {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Approve)

	var fr common.FederationRelationship
	// response.Status = common.FederationRelationshipStatusActive

	// if spireServerConsent and spireServerFederatedWithConsent == true => response.Status = common.FederationRelationshipStatusActive
	// fr.spireServerConsent = True
	return fr, errors.New("not implemented")
}

func (c *LocalHarvesterController) DenyFederationRelationship(ctx context.Context, spiffeID string) (common.FederationRelationship, error) {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.FederationRelationship, telemetry.Deny)

	var fr common.FederationRelationship
	// fr.spireServerConsent = False

	return fr, errors.New("not implemented")
}

func (c *LocalHarvesterController) GetFederationRelationshipsByStatus(ctx context.Context, status string) ([]common.FederationRelationship, error) {
	telemetry.Count(ctx, telemetry.HarvesterController, telemetry.FederationRelationship, telemetry.Get)

	var fr []common.FederationRelationship

	return fr, errors.New("not implemented")
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
