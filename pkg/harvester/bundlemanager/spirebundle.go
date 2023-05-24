package bundlemanager

import (
	"context"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/harvester/spireclient"
	"github.com/sirupsen/logrus"
)

// SpireBundleSyncer is responsible for periodically fetching the bundle from the SPIRE Server,
// signing it, and uploading it to the Galadriel Server.
type SpireBundleSyncer struct {
	spireClient     spireclient.Client
	galadrielClient galadrielclient.Client
	bundleSigner    integrity.Signer
	syncInterval    time.Duration
	logger          logrus.FieldLogger

	lastSpireBundle *spiffebundle.Bundle // last bundle fetched from the SPIRE Server and uploaded to the Galadriel Server
}

// SpireSyncerConfig holds the configuration for SpireBundleSyncer.
type SpireSyncerConfig struct {
	SpireClient     spireclient.Client
	GaladrielClient galadrielclient.Client
	BundleSigner    integrity.Signer
	SyncInterval    time.Duration
	Logger          logrus.FieldLogger
}

// NewSpireSyncer creates a new SpireBundleSyncer instance.
func NewSpireSyncer(config *SpireSyncerConfig) *SpireBundleSyncer {
	return &SpireBundleSyncer{
		spireClient:     config.SpireClient,
		galadrielClient: config.GaladrielClient,
		bundleSigner:    config.BundleSigner,
		syncInterval:    config.SyncInterval,
		logger:          config.Logger,
	}
}

// Run starts the SPIRE bundle syncer process.
func (s *SpireBundleSyncer) Run(ctx context.Context) error {
	s.logger.Info("SPIRE Bundle Syncer started")

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.syncSPIREBundle(ctx)
			if err != nil {
				s.logger.Errorf("Failed to sync SPIRE bundle: %v", err)
			}
		case <-ctx.Done():
			s.logger.Info("SPIRE Bundle Syncer stopped")
			return nil
		}
	}
}

// syncSPIREBundle periodically checks the SPIRE Server for a new bundle, signs it, and uploads the signed bundle to the Galadriel Server.
func (s *SpireBundleSyncer) syncSPIREBundle(ctx context.Context) error {
	s.logger.Debug("Checking SPIRE Server for a new bundle")

	spireCtx, cancel := context.WithTimeout(ctx, spireCallTimeout)
	defer cancel()

	// Fetch SPIRE bundle
	bundle, err := s.spireClient.GetBundle(spireCtx)
	if err != nil {
		return err
	}

	// Check if the bundle is the same as the last one fetched
	if s.lastSpireBundle != nil && s.lastSpireBundle.Equal(bundle) {
		return nil // No new bundle
	}

	s.logger.Debug("New bundle from SPIRE Server")

	// Generate the bundle to upload
	b, err := s.generateBundleToUpload(bundle)
	if err != nil {
		return fmt.Errorf("failed to create bundle to upload: %w", err)
	}

	galadrielCtx, cancel := context.WithTimeout(ctx, galadrielCallTimeout)
	defer cancel()

	// Upload the bundle to Galadriel Server
	err = s.galadrielClient.PostBundle(galadrielCtx, b)
	if err != nil {
		return fmt.Errorf("failed to upload bundle to Galadriel Server: %w", err)
	}

	s.logger.Info("Uploaded SPIRE bundle to Galadriel Server")
	s.lastSpireBundle = bundle
	return nil
}

// generateBundleToUpload creates an entity.Bundle using the provided SPIRE bundle.
// It marshals the SPIRE bundle, generates the bundle signature, and calculates the digest.
func (s *SpireBundleSyncer) generateBundleToUpload(spireBundle *spiffebundle.Bundle) (*entity.Bundle, error) {
	bundleBytes, err := spireBundle.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SPIRE bundle: %w", err)
	}

	bundleSignatureBytes, certChain, err := s.bundleSigner.Sign(bundleBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign the bundle: %w", err)
	}

	// Check if the signer returned a signing certificate to include in the bundle
	var cert []byte
	if len(certChain) > 0 {
		cert = certChain[0].Raw
	}

	digest := cryptoutil.CalculateDigest(bundleBytes)

	bundle := &entity.Bundle{
		Data:               bundleBytes,
		Digest:             digest[:],
		Signature:          bundleSignatureBytes,
		SigningCertificate: cert,
		TrustDomainName:    spireBundle.TrustDomain(),
	}

	return bundle, nil
}
