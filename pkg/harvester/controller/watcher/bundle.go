package watcher

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
)

var logger = logrus.WithField(telemetry.SubsystemName, telemetry.HarvesterController)

// BuildSelfBundleWatcher builds a watcher loop that periodically update the Galadriel Server
// with the latest root trust bundle from the domain (SPIRE Server) associated with this Harvester.
func BuildSelfBundleWatcher(interval time.Duration, server client.GaladrielServerClient, spire spire.SpireServer) util.RunnableTask {
	return func(ctx context.Context) error {
		t := time.NewTicker(interval)
		var currentDigest []byte

		for {
			select {
			case <-t.C:
				bundle, digest, hasNew := hasNewBundle(ctx, currentDigest, spire)
				if !hasNew {
					break
				}
				logger.Info("Bundle has changed, pushing to Galadriel Server")

				req, err := buildPostBundleRequest(bundle)
				if err != nil {
					logger.Error(err)
					break
				}

				if err = server.PostBundle(ctx, req); err != nil {
					logger.Errorf("Failed to push X.509 bundle: %v", err)
					break
				}
				logger.Debug("New bundle successfully pushed to Galadriel Server")

				currentDigest = digest
			case <-ctx.Done():
				return nil
			}
		}
	}
}

// BuildFederatedBundlesWatcher builds a watcher loop that periodically fetches the Galadriel Server
// for updates of any federation relationships this harvester is part of.
func BuildFederatedBundlesWatcher(interval time.Duration, server client.GaladrielServerClient, spire spire.SpireServer) util.RunnableTask {
	return func(ctx context.Context) error {
		t := time.NewTicker(interval)

		for {
			select {
			case <-t.C:
				req, err := buildSyncBundlesRequest(ctx, spire)
				if err != nil {
					logger.Errorf("Failed to build sync federated bundle request: %v", err)
					break
				}

				res, err := server.SyncFederatedBundles(ctx, req)
				if err != nil {
					logger.Errorf("Failed to get federated bundles updates: %v", err)
					break
				}

				bundles, processed := federatedBundlesUpdatesToSpiffeBundles(res)
				updatesLen := uint32(len(res.Updates))
				if updatesLen != processed {
					logger.Errorf("Failed to process %d out of %d trust domains", updatesLen-processed, updatesLen)
				}

				if len(bundles) == 0 {
					logger.Debug("No new federated bundles to set")
					break
				}

				logger.Infof("Setting %d new federated bundle(s)", len(bundles))
				if _, err = spire.SetFederatedBundles(ctx, bundles); err != nil {
					logger.Errorf("%v", err)
				}
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func hasNewBundle(ctx context.Context, currentDigest []byte, spire spire.SpireServer) (newBundle *spiffebundle.Bundle, newDigest []byte, hasNew bool) {
	spireBundle, err := spire.GetBundle(ctx)
	if err != nil {
		logger.Errorf("Failed to get spire bundle: %v", err)
		return nil, nil, false
	}

	b, err := spireBundle.X509Bundle().Marshal()
	if err != nil {
		logger.Errorf("Failed to marshal spire X.509 bundle: %v", err)
		return nil, nil, false
	}
	spireDigest := util.GetDigest(b)

	if !bytes.Equal(currentDigest, spireDigest) {
		return spireBundle, spireDigest, true
	}

	return nil, nil, false
}

func buildPostBundleRequest(b *spiffebundle.Bundle) (*common.PostBundleRequest, error) {
	bundle, err := b.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal X.509 bundle: %v", err)
	}

	x509b, err := b.X509Bundle().Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal X.509 bundle: %v", err)
	}

	ent := entity.Bundle{
		Data:            bundle,
		Digest:          util.GetDigest(x509b),
		TrustDomainName: b.TrustDomain(),
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
	}

	req := &common.PostBundleRequest{
		Bundle: &ent,
	}

	return req, nil
}

func buildSyncBundlesRequest(ctx context.Context, spire spire.SpireServer) (*common.SyncBundleRequest, error) {
	res, err := spire.GetFederatedBundles(ctx)
	if err != nil {
		return nil, err
	}

	digests := make(common.BundlesDigests)

	for _, b := range res.Bundles {
		td := b.TrustDomain()
		if err != nil {
			logger.Errorf("Failed to marshal bundle for trust domain %s: %v", td, err)
			continue
		}

		x509b, err := b.X509Bundle().Marshal()
		if err != nil {
			logger.Errorf("Failed to marshal X.509 bundle for trust domain %s: %v", td, err)
			continue
		}

		digests[td] = util.GetDigest(x509b)
	}

	state := &common.SyncBundleRequest{State: digests}

	return state, nil
}

func federatedBundlesUpdatesToSpiffeBundles(res *common.SyncBundleResponse) (bundles []*spiffebundle.Bundle, processed uint32) {
	for td, b := range res.Updates {
		if b.Data == nil {
			logger.Errorf("Received an empty bundle for trust domain %q", td)
			continue
		}

		bundle, err := spiffebundle.Parse(td, b.Data)
		if err != nil {
			logger.Errorf("Failed to parse trust bundle for %q: %v", td, err)
			continue
		}

		bundles = append(bundles, bundle)
		processed++
	}

	return bundles, processed
}
