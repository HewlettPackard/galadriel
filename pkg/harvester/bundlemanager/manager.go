package bundlemanager

import (
	"context"
	"errors"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
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
	federatedBundlesSyncer *FederatedSyncer
	spireBundleSyncer      *SpireBundleSyncer
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

	spireBundleSync := NewSpireSyncer(&SpireSyncerConfig{
		GaladrielClient: c.GaladrielClient,
		SpireClient:     c.SpireClient,
		BundleSigner:    c.BundleSigner,
		SyncInterval:    c.SpireBundlePollInterval,
		Logger:          c.Logger.WithField(telemetry.SubsystemName, telemetry.SpireBundleSyncer),
	})

	fedBundlesSync := NewFederatedSyncer(&FederatedSyncerConfig{
		GaladrielClient: c.GaladrielClient,
		SpireClient:     c.SpireClient,
		BundleVerifiers: c.BundleVerifiers,
		SyncInterval:    c.FederatedBundlesPollInterval,
		Logger:          c.Logger.WithField(telemetry.SubsystemName, telemetry.FederadBundlesSyncer),
	})

	return &BundleManager{
		federatedBundlesSyncer: fedBundlesSync,
		spireBundleSyncer:      spireBundleSync,
	}
}

// Run runs the bundle synchronization processes.
func (bm *BundleManager) Run(ctx context.Context) error {
	tasks := []func(ctx context.Context) error{
		bm.federatedBundlesSyncer.Run,
		bm.spireBundleSyncer.Run,
	}

	err := util.RunTasks(ctx, tasks...)
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
