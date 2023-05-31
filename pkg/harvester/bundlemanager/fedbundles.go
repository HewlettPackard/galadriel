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

// FederatedBundlesSynchronizer is responsible for periodically synchronizing the federated bundles
// in the SPIRE Server with those fetched from the Galadriel Server. The synchronization process consists of the following steps:
// 1. Fetch the federated bundles from the Galadriel Server.
// 2. Verify the integrity of these bundles using the provided bundle verifiers.
// 3. Update the SPIRE Server with the new bundles.
// 4. If any relationships no longer exist, remove the corresponding bundles from the SPIRE Server.
//
// The removal of bundles is done in DISSOCIATE mode, which dissociates the registration entries
// from the non-existent federated trust domains. It also maintains a last-known state of federated
// bundles fetched from the Galadriel Server to optimize synchronizations.
type FederatedBundlesSynchronizer struct {
	spireClient     spireclient.Client
	galadrielClient galadrielclient.Client
	bundleVerifiers []integrity.Verifier
	syncInterval    time.Duration
	logger          logrus.FieldLogger

	// last state of Federated Bundles fetched from Galadriel Server
	lastFederatedBundleDigests map[spiffeid.TrustDomain][]byte
}

// FederatedBundlesSynchronizerConfig holds the configuration for FederatedBundlesSynchronizer.
type FederatedBundlesSynchronizerConfig struct {
	SpireClient     spireclient.Client
	GaladrielClient galadrielclient.Client
	BundleVerifiers []integrity.Verifier
	SyncInterval    time.Duration
	Logger          logrus.FieldLogger
}

func NewFederatedBundlesSynchronizer(config *FederatedBundlesSynchronizerConfig) *FederatedBundlesSynchronizer {
	return &FederatedBundlesSynchronizer{
		spireClient:     config.SpireClient,
		galadrielClient: config.GaladrielClient,
		bundleVerifiers: config.BundleVerifiers,
		syncInterval:    config.SyncInterval,
		logger:          config.Logger,
	}
}

// StartSyncing starts the synchronization process.
func (s *FederatedBundlesSynchronizer) StartSyncing(ctx context.Context) error {
	s.logger.Info("Federated Bundles Synchronizer started")

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.synchronizeFederatedBundles(ctx); err != nil {
				s.logger.Errorf("Failed to sync federated bundles with Galadriel Server: %v", err)
			}
		case <-ctx.Done():
			s.logger.Info("Federated Bundles Synchronizer stopped")
			return nil
		}
	}
}

func (s *FederatedBundlesSynchronizer) synchronizeFederatedBundles(ctx context.Context) error {
	s.logger.Debug("Synchronize federated bundles with Galadriel Server")

	spireCallCtx, spireCallCancel := context.WithTimeout(ctx, spireCallTimeout)
	if spireCallCancel == nil {
		return fmt.Errorf("failed to create context for SPIRE call")
	}
	defer spireCallCancel()

	fedBundlesInSPIRE, err := s.fetchSPIREFederatedBundles(spireCallCtx)
	if err != nil {
		return fmt.Errorf("failed to fetch federated bundles from SPIRE Server: %w", err)
	}

	galadrielCallCtx, galadrielCallCancel := context.WithTimeout(ctx, galadrielCallTimeout)
	if galadrielCallCancel == nil {
		return fmt.Errorf("failed to create context for Galadriel call")
	}
	defer galadrielCallCancel()

	bundles, digests, err := s.galadrielClient.SyncBundles(galadrielCallCtx, fedBundlesInSPIRE)
	if err != nil {
		return fmt.Errorf("failed to sync federated bundles with Galadriel Server: %w", err)
	}

	// if the federated bundles have not changed since last server poll, skip the sync
	if areMapsEqual(s.lastFederatedBundleDigests, digests) {
		s.logger.Debug("Federated bundles have not changed")
		return nil
	}

	bundlesToSet := make([]*spiffebundle.Bundle, 0)
	for _, b := range bundles {
		if err := s.validateBundleIntegrity(b); err != nil {
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

	setStatuses, err := s.spireClient.SetFederatedBundles(spireCallCtx, bundlesToSet)
	if err != nil {
		s.logger.Errorf("Failed to set federated bundles in SPIRE Server: %v", err)
	} else {
		s.logFederatedBundleSetStatuses(setStatuses)
	}

	bundlesToDelete := s.findTrustDomainsToDelete(fedBundlesInSPIRE, digests)
	if len(bundlesToDelete) == 0 {
		// No updates to be made, update the last state and return
		s.lastFederatedBundleDigests = digests
		return nil
	}

	deleteStatuses, err := s.spireClient.DeleteFederatedBundles(spireCallCtx, bundlesToDelete)
	if err != nil {
		s.logger.Errorf("Failed to delete federated bundles in SPIRE Server: %v", err)
	} else {
		s.logFederatedBundleDeleteStatuses(deleteStatuses)
	}

	// update the last state of federated bundles
	s.lastFederatedBundleDigests = digests

	return nil
}

// findTrustDomainsToDelete returns a slice of trust domains to delete based on the provided bundles and digests map.
func (s *FederatedBundlesSynchronizer) findTrustDomainsToDelete(bundles []*entity.Bundle, digests map[spiffeid.TrustDomain][]byte) []spiffeid.TrustDomain {
	trustDomainsToDelete := make([]spiffeid.TrustDomain, 0)
	for _, b := range bundles {
		if _, ok := digests[b.TrustDomainName]; !ok {
			trustDomainsToDelete = append(trustDomainsToDelete, b.TrustDomainName)
		}
	}

	return trustDomainsToDelete
}

// validateBundleIntegrity verifies the bundle using the given verifiers.
// If one of the verifiers can verify the bundle, it returns nil.
func (s *FederatedBundlesSynchronizer) validateBundleIntegrity(bundle *entity.Bundle) error {
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

func (s *FederatedBundlesSynchronizer) fetchSPIREFederatedBundles(ctx context.Context) ([]*entity.Bundle, error) {
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

func (s *FederatedBundlesSynchronizer) logFederatedBundleSetStatuses(federatedBundleStatuses []*spireclient.BatchSetFederatedBundleStatus) {
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

func (s *FederatedBundlesSynchronizer) logFederatedBundleDeleteStatuses(deleteStatuses []*spireclient.BatchDeleteFederatedBundleStatus) {
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

func areMapsEqual(map1, map2 map[spiffeid.TrustDomain][]byte) bool {
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
