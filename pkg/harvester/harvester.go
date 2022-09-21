package harvester

import (
	"context"
	"errors"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
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

// Run starts running the Harvester, starting its endpoints.
func (h *Harvester) Run(ctx context.Context) error {
	if err := h.run(ctx); err != nil {
		return err
	}
	return nil
}

func (h *Harvester) run(ctx context.Context) (err error) {
	cat, err := catalog.Load(ctx, catalog.Config{Log: h.config.Log})
	if err != nil {
		return err
	}
	defer cat.Close()

	config := &controller.Config{
		ServerAddress:   h.config.ServerAddress,
		SpireSocketPath: h.config.SpireAddress,
		Log:             h.config.Log.WithField(telemetry.SubsystemName, telemetry.HarvesterController),
		Metrics:         h.config.metrics,
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
