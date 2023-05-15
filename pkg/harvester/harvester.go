package harvester

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
)

const jwtRotationInterval = 12 * time.Hour

// Harvester represents a Galadriel Harvester
type Harvester struct {
	controller controller.HarvesterController //nolint:unused
	client     client.GaladrielServerClient
	config     *Config
}

// New creates a new instances of Harvester with the given configuration.
func New(config *Config) *Harvester {
	return &Harvester{
		config: config,
	}
}

// Run starts running the Harvester.
func (h *Harvester) Run(ctx context.Context) error {
	h.config.Logger.Info("Starting Harvester")

	if h.config.JoinToken == "" {
		return errors.New("token is required to connect the Harvester to the Galadriel Server")
	}

	galadrielClient, err := client.NewGaladrielServerClient(h.config.ServerAddress, h.config.ServerTrustBundlePath)
	if err != nil {
		return fmt.Errorf("failed to create Galadriel Server client: %w", err)
	}
	h.client = galadrielClient

	err = galadrielClient.Onboard(ctx, h.config.JoinToken)
	if err != nil {
		return fmt.Errorf("failed to connect to Galadriel Server: %w", err)
	}

	config := &controller.Config{
		GaladrielServerClient: galadrielClient,
		SpireSocketPath:       h.config.SpireAddress,
		BundleUpdatesInterval: h.config.BundleUpdatesInterval,
		Logger:                h.config.Logger.WithField(telemetry.SubsystemName, telemetry.HarvesterController),
	}
	c, err := controller.NewHarvesterController(ctx, config)
	if err != nil {
		return err
	}

	err = util.RunTasks(ctx, c.Run, h.startJWTTokenRotation)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (h *Harvester) startJWTTokenRotation(ctx context.Context) error {
	h.config.Logger.Info("Starting JWT token rotator")

	ticker := time.NewTicker(jwtRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.config.Logger.Info("Requesting a new JWT token from Galadriel Server")
			err := h.client.GetNewJWTToken(ctx)
			if err != nil {
				h.config.Logger.Errorf("Error getting new JWT token: %v", err)
				return err
			}
		case <-ctx.Done():
			h.config.Logger.Info("Stopped JWT token rotator")
			return nil
		}
	}
}
