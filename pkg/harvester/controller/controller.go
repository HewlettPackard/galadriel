package controller

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller/watcher"
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
	config *Config
}

// Config represents the configurations for the Harvester Controller
type Config struct {
	ServerAddress         string
	SpireSocketPath       net.Addr
	RootCAPath            string
	JoinToken             string
	BundleUpdatesInterval time.Duration
	Logger                logrus.FieldLogger
}

func NewHarvesterController(ctx context.Context, config *Config) (*HarvesterController, error) {
	sc := spire.NewLocalSpireServer(ctx, config.SpireSocketPath)
	gc, err := client.NewGaladrielServerClient(config.ServerAddress, config.JoinToken, config.RootCAPath)
	if err != nil {
		return nil, err
	}

	// Authenticates with join token
	// TODO: Attempt to authenticate with SVID if missing token
	err = gc.Connect(ctx, config.JoinToken)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection with Galadriel Server: %w", err)
	}

	return &HarvesterController{
		spire:  sc,
		server: gc,
		config: config,
		logger: logrus.WithField(telemetry.SubsystemName, telemetry.HarvesterController),
	}, nil
}

func (c *HarvesterController) Run(ctx context.Context) error {
	c.logger.Info("Starting harvester controller")

	go c.run(ctx)

	<-ctx.Done()
	c.logger.Debug("Shutting down...")

	return nil
}

func (c *HarvesterController) run(ctx context.Context) {
	federatedBundlesInterval := time.Second * 10

	err := util.RunTasks(ctx,
		watcher.BuildSelfBundleWatcher(c.config.BundleUpdatesInterval, c.server, c.spire),
		watcher.BuildFederatedBundlesWatcher(federatedBundlesInterval, c.server, c.spire),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		c.logger.Error(err)
	}
}
