package harvester

import (
	"context"
	"errors"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
)

// Harvester represents a Galadriel Harvester
type Harvester struct {
	controller controller.HarvesterController //nolint:unused
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

	config := &controller.Config{
		ServerAddress:         h.config.ServerAddress,
		RootCAPath:            h.config.RootCAPath,
		SpireSocketPath:       h.config.SpireAddress,
		JoinToken:             h.config.JoinToken,
		BundleUpdatesInterval: h.config.BundleUpdatesInterval,
		Logger:                h.config.Logger.WithField(telemetry.SubsystemName, telemetry.HarvesterController),
	}
	c, err := controller.NewHarvesterController(ctx, config)
	if err != nil {
		return err
	}

	err = util.RunTasks(ctx, c.Run)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (h *Harvester) Stop() {
	// unload and cleanup stuff
}
