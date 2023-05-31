package bundlemanager

import (
	"context"
	"fmt"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spireclient"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
)

// SpireBundleSynchronizer manages the synchronization of bundles from the SPIRE server.
// It periodically fetches the bundle from the SPIRE Server, signs it, and uploads it to the Galadriel Server.
type SpireBundleSynchronizer struct {
	spireClient     spireclient.Client
	galadrielClient galadrielclient.Client
	bundleSigner    integrity.Signer
	syncInterval    time.Duration
	logger          logrus.FieldLogger

	lastSpireBundle *spiffebundle.Bundle // last bundle fetched from the SPIRE Server and uploaded to the Galadriel Server
}

// SpireSynchronizerConfig holds the configuration for SpireBundleSynchronizer.
type SpireSynchronizerConfig struct {
	SpireClient     spireclient.Client
	GaladrielClient galadrielclient.Client
	BundleSigner    integrity.Signer
	SyncInterval    time.Duration
	Logger          logrus.FieldLogger
}

// NewSpireSynchronizer creates a new SpireBundleSynchronizer instance.
func NewSpireSynchronizer(config *SpireSynchronizerConfig) *SpireBundleSynchronizer {
	return &SpireBundleSynchronizer{
		spireClient:     config.SpireClient,
		galadrielClient: config.GaladrielClient,
		bundleSigner:    config.BundleSigner,
		syncInterval:    config.SyncInterval,
		logger:          config.Logger,
	}
}

// StartSyncing initializes the SPIRE bundle synchronization process.
// It starts an infinite loop that periodically fetches the SPIRE bundle, signs it and uploads it to the Galadriel Server.
func (s *SpireBundleSynchronizer) StartSyncing(ctx context.Context) error {
	s.logger.Info("SPIRE Bundle Synchronizer started")

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.syncSpireBundleToGaladriel(ctx)
			if err != nil {
				s.logger.Errorf("Failed to sync SPIRE bundle: %v", err)
			}
		case <-ctx.Done():
			s.logger.Info("SPIRE Bundle Synchronizer stopped")
			return nil
		}
	}
}

// syncSPIREBundle fetches a new bundle from the SPIRE Server,
// signs it if it's new, and then uploads it to the Galadriel Server.
func (s *SpireBundleSynchronizer) syncSpireBundleToGaladriel(ctx context.Context) error {
	s.logger.Debug("Checking SPIRE Server for a new bundle")

	spireCallCtx, spireCallCancel := context.WithTimeout(ctx, spireCallTimeout)
	if spireCallCancel == nil {
		return fmt.Errorf("failed to create context for SPIRE call")
	}
	defer spireCallCancel()

	// Fetch SPIRE bundle
	bundle, err := s.spireClient.GetBundle(spireCallCtx)
	if err != nil {
		return err
	}

	// Check if the bundle is the same as the last one fetched
	if s.lastSpireBundle != nil && s.lastSpireBundle.Equal(bundle) {
		return nil // No new bundle
	}

	s.logger.Debug("New bundle from SPIRE Server")

	// Generate the bundle to upload
	bundleToUpload, err := s.prepareBundleForUpload(bundle)
	if err != nil {
		return fmt.Errorf("failed to create bundle to upload: %w", err)
	}

	galadrielCallCtx, galadrielCallCancel := context.WithTimeout(ctx, galadrielCallTimeout)
	if galadrielCallCancel == nil {
		return fmt.Errorf("failed to create context for Galadriel Server call")
	}
	defer galadrielCallCancel()

	// Upload the bundle to Galadriel Server
	err = s.galadrielClient.PostBundle(galadrielCallCtx, bundleToUpload)
	if err != nil {
		return fmt.Errorf("failed to upload bundle to Galadriel Server: %w", err)
	}

	s.logger.Info("Uploaded SPIRE bundle to Galadriel Server")
	s.lastSpireBundle = bundle
	return nil
}

// prepareBundleForUpload creates an entity.Bundle using the provided SPIRE bundle.
// It marshals the SPIRE bundle, generates the bundle signature, and calculates the digest.
func (s *SpireBundleSynchronizer) prepareBundleForUpload(spireBundle *spiffebundle.Bundle) (*entity.Bundle, error) {
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
