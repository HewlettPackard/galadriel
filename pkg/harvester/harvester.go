package harvester

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/common/util/fileutil"
	"github.com/HewlettPackard/galadriel/pkg/harvester/bundlemanager"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
	"github.com/HewlettPackard/galadriel/pkg/harvester/endpoints"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spireclient"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// Harvester represents the Harvester agent.
// It starts the bundle manager and the endpoints.
type Harvester struct {
	c *Config
}

// Config conveys the configuration of the Harvester.
type Config struct {
	TrustDomain                  spiffeid.TrustDomain
	HarvesterSocketPath          net.Addr     // UDS socket address the Harvester will listen on
	SpireSocketPath              net.Addr     // UDS socket address the SPIRE server listens on and Harvester will connect to
	GaladrielServerAddress       *net.TCPAddr // TCP address the Galadriel Server listens on and Harvester will connect to
	JoinToken                    string
	BundleUpdatesInterval        time.Duration
	FederatedBundlesPollInterval time.Duration
	SpireBundlePollInterval      time.Duration
	ServerTrustBundlePath        string
	DataDir                      string
	Logger                       logrus.FieldLogger
	ProvidersConfig              *catalog.ProvidersConfig
}

func New(cfg *Config) *Harvester {
	return &Harvester{
		c: cfg,
	}
}

// Run starts the Harvester and orchestrates the main functionality.
// It performs the following steps:
// - Loads catalogs from the providers configuration.
// - Creates the data directory if it does not exist.
// - Creates a client for Galadriel Server.
// - Onboards the Harvester to Galadriel Server if it is not already onboarded.
// - Creates a SPIRE client using the provided SPIRE address.
// - Creates, configures and run the Harvester endpoints.
// - Creates and runs the BundleManager responsible for bundles synchronization.
func (h *Harvester) Run(ctx context.Context) error {
	h.c.Logger.Info("Starting Harvester")

	cat := catalog.New()
	err := cat.LoadFromProvidersConfig(h.c.ProvidersConfig)
	if err != nil {
		return fmt.Errorf("failed to load catalogs from providers config: %w", err)
	}

	err = fileutil.CreateDirIfNotExist(h.c.DataDir)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	galadrielClient, err := galadrielclient.NewClient(ctx, &galadrielclient.Config{
		TrustDomain:            h.c.TrustDomain,
		GaladrielServerAddress: h.c.GaladrielServerAddress,
		TrustBundlePath:        h.c.ServerTrustBundlePath,
		DataDir:                h.c.DataDir,
		JoinToken:              h.c.JoinToken,
		Logger:                 h.c.Logger.WithField(telemetry.SubsystemName, telemetry.Harvester),
	})
	if err != nil {
		h.c.Logger.Error("Harvester could not connect to Server. Needs to be re-onboarded with new join token")
		return fmt.Errorf("failed to create Galadriel Server client: %w", err)
	}

	spireClient, err := spireclient.NewSpireClient(ctx, h.c.SpireSocketPath)
	if err != nil {
		return fmt.Errorf("failed to create SPIRE client: %w", err)
	}

	ep, err := endpoints.New(&endpoints.Config{
		LocalAddress: h.c.HarvesterSocketPath,
		Client:       galadrielClient,
		Logger:       h.c.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
	})
	if err != nil {
		return fmt.Errorf("failed to create Harvester endpoints: %w", err)
	}

	bundleManager := bundlemanager.NewBundleManager(&bundlemanager.Config{
		SpireClient:                  spireClient,
		GaladrielClient:              galadrielClient,
		FederatedBundlesPollInterval: h.c.FederatedBundlesPollInterval,
		SpireBundlePollInterval:      h.c.SpireBundlePollInterval,
		BundleSigner:                 cat.GetBundleSigner(),
		BundleVerifiers:              cat.GetBundleVerifiers(),
		Logger:                       h.c.Logger,
	})

	tasks := []func(ctx context.Context) error{
		ep.ListenAndServe,
		bundleManager.Run,
	}

	err = util.RunTasks(ctx, tasks...)
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
