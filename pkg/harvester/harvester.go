package harvester

import (
	"context"
	"errors"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
)

// Harvester represents a Galadriel Harvester
type Harvester struct {
	controller controller.HarvesterController //nolint:unused
	api        api.API                        //nolint:unused

	config *Config
}

// New creates a new instances of Harvester with the given configuration.
func New(config *Config) *Harvester {
	return &Harvester{
		config: config,
	}
}

// Run starts running the Harvester.
func (h *Harvester) Run(ctx context.Context) error {
	h.config.Log.Info("Starting Harvester")

	if h.config.AccessToken == "" {
		return errors.New("token is required to connect the Harvester to the Galadriel Server")
	}

	galadrielClient, err := client.NewGaladrielServerClient(h.config.ServerAddress, h.config.AccessToken)
	if err != nil {
		return err
	}

	err = galadrielClient.Connect(ctx, h.config.AccessToken)
	if err != nil {
		return err
	}

	config := &controller.Config{
		ServerAddress:         h.config.ServerAddress,
		SpireSocketPath:       h.config.SpireAddress,
		AccessToken:           h.config.AccessToken,
		BundleUpdatesInterval: h.config.BundleUpdatesInterval,
		Log:                   h.config.Log.WithField(telemetry.SubsystemName, telemetry.HarvesterController),
		Metrics:               h.config.metrics,
	}
	c, err := controller.NewHarvesterController(ctx, config)
	if err != nil {
		return err
	}

	tasks := []func(context.Context) error{
		c.Run,
	}

	err = util.RunTasks(ctx, tasks)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (h *Harvester) Stop() {
	// unload and cleanup stuff
}
