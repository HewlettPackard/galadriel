package bundlemanager

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/HewlettPackard/galadriel/pkg/harvester/models"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spireclient"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/grpc/codes"
)

// FederatedSynchronizer is responsible for periodically syncing the federated bundles in SPIRE Server
// with the Federated bundles fetched from Galadriel Server.
// This process involves:
// 1. Fetching the federated bundles from Galadriel Server.
// 2. Verifying the integrity of the bundles.
// 3. Setting the new bundles to SPIRE Server.
// 4. Deleting the bundles from SPIRE Server for relationships that no longer exist.
// The deletion of bundles in SPIRE Server is done using the DISSOCIATE mode deletes the federated bundles
// dissociating the registration entries from the federated trust domain.
type FederatedSynchronizer struct {
	spireClient     spireclient.Client
	galadrielClient galadrielclient.Client
	bundleVerifiers []integrity.Verifier
	syncInterval    time.Duration
	logger          logrus.FieldLogger

	// last state of Federated Bundles fetched from Galadriel Server
	lastFederatesBundlesDigests map[spiffeid.TrustDomain][]byte
}

// FederatedSynchronizerConfig holds the configuration for FederatedSynchronizer.
type FederatedSynchronizerConfig struct {
	SpireClient     spireclient.Client
	GaladrielClient galadrielclient.Client
	BundleVerifiers []integrity.Verifier
	SyncInterval    time.Duration
	Logger          logrus.FieldLogger
}

func NewFederatedSynchronizer(config *FederatedSynchronizerConfig) *FederatedSynchronizer {
	return &FederatedSynchronizer{
		spireClient:     config.SpireClient,
		galadrielClient: config.GaladrielClient,
		bundleVerifiers: config.BundleVerifiers,
		syncInterval:    config.SyncInterval,
		logger:          config.Logger,
	}
}

// Run starts the synchronization process.
func (s *FederatedSynchronizer) Run(ctx context.Context) error {
	s.logger.Info("Federated Bundles Synchronizer started")

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.syncFederatedBundles(ctx); err != nil {
				s.logger.Errorf("Failed to sync federated bundles with Galadriel Server: %v", err)
			}
		case <-ctx.Done():
			s.logger.Info("Federated Bundles Synchronizer stopped")
			return nil
		}
	}
}

func (s *FederatedSynchronizer) syncFederatedBundles(ctx context.Context) error {
	s.logger.Debug("Syncing federated bundles with Galadriel Server")

	spireCtx, spireCancel := context.WithTimeout(ctx, spireCallTimeout)
	if spireCancel == nil {
		return fmt.Errorf("failed to create context for SPIRE call")
	}
	defer spireCancel()

	fedBundlesInSPIRE, err := s.fetchSPIREFederatedBundles(spireCtx)
	if err != nil {
		return fmt.Errorf("failed to fetch federated bundles from SPIRE Server: %w", err)
	}

	galadrielCtx, galadrielCancel := context.WithTimeout(ctx, galadrielCallTimeout)
	if galadrielCancel == nil {
		return fmt.Errorf("failed to create context for Galadriel call")
	}
	defer galadrielCancel()

	bundles, digests, err := s.galadrielClient.SyncBundles(galadrielCtx, fedBundlesInSPIRE)
	if err != nil {
		return fmt.Errorf("failed to sync federated bundles with Galadriel Server: %w", err)
	}

	// if the federated bundles have not changed since last server poll, skip the sync
	if equalMaps(s.lastFederatesBundlesDigests, digests) {
		s.logger.Debug("Federated bundles have not changed")
		return nil
	}

	bundlesToSet := make([]*spiffebundle.Bundle, 0)
	for _, b := range bundles {
		if err := s.verifyBundle(b); err != nil {
			s.logger.Errorf("Failed to verify bundle for trust domain %q: %v", b.TrustDomainName, err)
			continue // skip the bundle
		}

		spireBundle, err := models.ConvertEntityBundleToSPIFFEBundle(b)
		if err != nil {
			s.logger.Errorf("failed to convert bundle for trust domain %q: %v", b.TrustDomainName, err)
			continue // skip the bundle
		}

		bundlesToSet = append(bundlesToSet, spireBundle)
	}

	setStatuses, err := s.spireClient.SetFederatedBundles(spireCtx, bundlesToSet)
	if err != nil {
		s.logger.Errorf("Failed to set federated bundles in SPIRE Server: %v", err)
	} else {
		s.logFederatedBundleSetStatuses(setStatuses)
	}

	bundlesToDelete := s.getTrustDomainsToDelete(fedBundlesInSPIRE, digests)
	if len(bundlesToDelete) == 0 {
		// No updates to be made, update the last state and return
		s.lastFederatesBundlesDigests = digests
		return nil
	}

	deleteStatuses, err := s.spireClient.DeleteFederatedBundles(spireCtx, bundlesToDelete)
	if err != nil {
		s.logger.Errorf("Failed to delete federated bundles in SPIRE Server: %v", err)
	} else {
		s.logFederatedBundleDeleteStatuses(deleteStatuses)
	}

	// update the last state of federated bundles
	s.lastFederatesBundlesDigests = digests

	return nil
}

// getTrustDomainsToDelete returns a slice of trust domains to delete based on the provided bundles and digests map.
func (s *FederatedSynchronizer) getTrustDomainsToDelete(bundles []*entity.Bundle, digests map[spiffeid.TrustDomain][]byte) []spiffeid.TrustDomain {
	trustDomainsToDelete := make([]spiffeid.TrustDomain, 0)
	for _, b := range bundles {
		if _, ok := digests[b.TrustDomainName]; !ok {
			trustDomainsToDelete = append(trustDomainsToDelete, b.TrustDomainName)
		}
	}

	return trustDomainsToDelete
}

// verifyBundle verifies the bundle using the given verifiers.
// If one of the verifiers can verify the bundle, it returns nil.
func (s *FederatedSynchronizer) verifyBundle(bundle *entity.Bundle) error {
	var certChain []*x509.Certificate
	if len(bundle.SigningCertificate) > 0 {
		var err error
		certChain, err = x509.ParseCertificates(bundle.SigningCertificate)
		if err != nil {
			return fmt.Errorf("failed to parse signing certificate chain: %w", err)
		}
	}

	for _, verifier := range s.bundleVerifiers {
		err := verifier.Verify(bundle.Data, bundle.Signature, certChain)
		if err == nil {
			return nil
		}
		s.logger.Warnf("Bundle for trust domain %q failed verification using %T verifier: %v ", bundle.TrustDomainName, verifier, err)
	}

	return fmt.Errorf("no verifier could verify the bundle")
}

func (s *FederatedSynchronizer) fetchSPIREFederatedBundles(ctx context.Context) ([]*entity.Bundle, error) {
	bundles, err := s.spireClient.GetFederatedBundles(ctx)
	if err != nil {
		return nil, err
	}

	entBundles := make([]*entity.Bundle, 0, len(bundles))
	for _, b := range bundles {
		entB, err := models.ConvertSPIFFEBundleToEntityBundle(b)
		if err != nil {
			return nil, err
		}
		entBundles = append(entBundles, entB)
	}

	return entBundles, nil
}

func (s *FederatedSynchronizer) logFederatedBundleSetStatuses(federatedBundleStatuses []*spireclient.BatchSetFederatedBundleStatus) {
	for _, status := range federatedBundleStatuses {
		if status.Status.Code != codes.OK {
			s.logger.WithFields(logrus.Fields{
				telemetry.TrustDomain:    status.Bundle.TrustDomain(),
				telemetry.BundleOpStatus: status.Status.Message,
			}).Error("Failed setting federated bundle", status.Bundle.TrustDomain(), status.Status)
		} else {
			s.logger.WithField(telemetry.TrustDomain, status.Bundle.TrustDomain()).Info("Federated bundle set")
		}
	}
}

func (s *FederatedSynchronizer) logFederatedBundleDeleteStatuses(deleteStatuses []*spireclient.BatchDeleteFederatedBundleStatus) {
	for _, status := range deleteStatuses {
		if status.Status.Code != codes.OK {
			s.logger.WithFields(logrus.Fields{
				telemetry.TrustDomain:    status.TrustDomain,
				telemetry.BundleOpStatus: status.Status.Message,
			}).Error("Failed deleting federated bundle", status.TrustDomain, status.Status)
		} else {
			s.logger.WithField(telemetry.TrustDomain, status.TrustDomain).Info("Federated bundle deleted")
		}
	}
}

func equalMaps(map1, map2 map[spiffeid.TrustDomain][]byte) bool {
	if len(map1) == 0 && len(map2) == 0 {
		return true
	}
	if len(map1) != len(map2) {
		return false
	}
	for domain, value1 := range map1 {
		value2, ok := map2[domain]
		if !ok || !bytes.Equal(value1, value2) {
			return false
		}
	}
	return true
}
