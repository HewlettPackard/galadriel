package controller

import (
	"context"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
	"github.com/sirupsen/logrus"
)

// HarvesterController represents the component responsible for handling
// the controller loops that will keep sending fresh bundles and configurations
// to and from SPIRE Server and Galadriel Server.
type HarvesterController struct {
	logger logrus.FieldLogger
	spire  spire.SpireServer
	server client.GaladrielServerClient
}

// Config represents the configurations for the Harvester Controller
type Config struct {
	ServerAddress   string
	SpireSocketPath net.Addr
	Log             logrus.FieldLogger
	Metrics         telemetry.MetricServer
}

func NewHarvesterController(ctx context.Context, config *Config) (*HarvesterController, error) {
	sc := spire.NewLocalSpireServer(ctx, config.SpireSocketPath)
	gc, err := client.NewGaladrielServerClient(config.ServerAddress)
	if err != nil {
		return nil, err
	}

	return &HarvesterController{
		spire:  sc,
		server: gc,
		logger: logrus.WithField(telemetry.SubsystemName, telemetry.HarvesterController),
	}, nil
}

func (c *HarvesterController) Run(ctx context.Context) error {
	c.logger.Info("Starting harvester controller")

	go c.run(ctx)

	<-ctx.Done()
	return nil
}

func (c *HarvesterController) run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			telemetry.Count(ctx, telemetry.HarvesterController, telemetry.TrustBundle, telemetry.Add)
		case <-ctx.Done():
			c.logger.Debug("Done")
			return
		}
	}
}
