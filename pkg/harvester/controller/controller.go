package controller

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
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
	AccessToken           string
	BundleUpdatesInterval time.Duration
	Log                   logrus.FieldLogger
	Metrics               telemetry.MetricServer
}

func NewHarvesterController(ctx context.Context, config *Config) (*HarvesterController, error) {
	sc := spire.NewLocalSpireServer(ctx, config.SpireSocketPath)
	gc, err := client.NewGaladrielServerClient(config.ServerAddress, config.AccessToken)
	if err != nil {
		return nil, err
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
	err := util.RunTasks(ctx, []func(context.Context) error{
		c.notifyBundleUpdates,
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		c.logger.Error(err)
	}
}

func (c *HarvesterController) notifyBundleUpdates(ctx context.Context) error {
	t := time.NewTicker(c.config.BundleUpdatesInterval)
	var currentBundle *spiffebundle.Bundle

	for {
		select {
		case <-t.C:
			b, hasNew := c.hasNewBundle(ctx, currentBundle)
			if hasNew {
				c.logger.Info("Bundle has changed, pushing to Galadriel")

				x509b, err := b.X509Bundle().Marshal()
				if err != nil {
					c.logger.Error("failed to marshal X.509 bundle: %v", err)
				}

				err = c.server.PostBundle(ctx, &common.PostBundleRequest{
					TrustBundle: common.TrustBundle{
						TrustDomain:  b.TrustDomain(),
						Bundle:       x509b,
						BundleDigest: util.GetDigest(x509b),
					},
				})
				if err != nil {
					c.logger.Errorf("failed to push X.509 bundle: %v", err)
					break
				}

				currentBundle = b
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *HarvesterController) hasNewBundle(ctx context.Context, current *spiffebundle.Bundle) (*spiffebundle.Bundle, bool) {
	b, err := c.spire.GetBundle(ctx)
	if err != nil {
		c.logger.Errorf("failed to check bundle updates: %v", err)
		return nil, false
	}

	if !current.Equal(b) {
		return b, true
	}

	return nil, false
}
