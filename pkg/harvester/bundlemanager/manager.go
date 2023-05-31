package bundlemanager

import (
	"context"
	"errors"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spireclient"
	"github.com/sirupsen/logrus"
)

const (
	defaultFederatedBundlesPollInterval = 2 * time.Minute
	defaultSpireBundlesPollInterval     = 1 * time.Minute
	spireCallTimeout                    = 10 * time.Second
	galadrielCallTimeout                = 2 * time.Minute
)

// BundleManager is responsible for managing the synchronization and watching of bundles.
type BundleManager struct {
	federatedBundlesSynchronizer *FederatedBundlesSynchronizer
	spireBundleSynchronizer      *SpireBundleSynchronizer
}

// Config holds the configuration for BundleManager.
type Config struct {
	SpireClient                  spireclient.Client
	GaladrielClient              galadrielclient.Client
	FederatedBundlesPollInterval time.Duration
	SpireBundlePollInterval      time.Duration

	// BundleSigner is used to sign the bundle before uploading it to Galadriel Server.
	BundleSigner integrity.Signer
	// BundleVerifiers are used to verify the bundle received from the SPIRE Server.
	BundleVerifiers []integrity.Verifier

	Logger logrus.FieldLogger
}

// NewBundleManager creates a new BundleManager instance.
func NewBundleManager(c *Config) *BundleManager {
	if c.FederatedBundlesPollInterval == 0 {
		c.FederatedBundlesPollInterval = defaultFederatedBundlesPollInterval
	}
	if c.SpireBundlePollInterval == 0 {
		c.SpireBundlePollInterval = defaultSpireBundlesPollInterval
	}

	spireBundleSync := NewSpireSynchronizer(&SpireSynchronizerConfig{
		GaladrielClient: c.GaladrielClient,
		SpireClient:     c.SpireClient,
		BundleSigner:    c.BundleSigner,
		SyncInterval:    c.SpireBundlePollInterval,
		Logger:          c.Logger.WithField(telemetry.SubsystemName, telemetry.SpireBundleSynchronizer),
	})

	fedBundlesSync := NewFederatedBundlesSynchronizer(&FederatedBundlesSynchronizerConfig{
		GaladrielClient: c.GaladrielClient,
		SpireClient:     c.SpireClient,
		BundleVerifiers: c.BundleVerifiers,
		SyncInterval:    c.FederatedBundlesPollInterval,
		Logger:          c.Logger.WithField(telemetry.SubsystemName, telemetry.FederatedBundlesSynchronizer),
	})

	return &BundleManager{
		federatedBundlesSynchronizer: fedBundlesSync,
		spireBundleSynchronizer:      spireBundleSync,
	}
}

// Run runs the bundle synchronization processes.
func (bm *BundleManager) Run(ctx context.Context) error {
	tasks := []func(ctx context.Context) error{
		bm.federatedBundlesSynchronizer.StartSyncing,
		bm.spireBundleSynchronizer.StartSyncing,
	}

	err := util.RunTasks(ctx, tasks...)
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
