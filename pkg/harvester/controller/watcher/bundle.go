package watcher

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/client"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
)

var logger = logrus.WithField(telemetry.SubsystemName, telemetry.HarvesterController)

func BuildSelfBundleWatcher(interval time.Duration, server client.GaladrielServerClient, spire spire.SpireServer) util.RunnableTask {
	return func(ctx context.Context) error {
		t := time.NewTicker(interval)
		var (
			currentDigest []byte
			// harvesterID   string
		)

		for {
			select {
			case <-t.C:
				var err error
				//if ds != nil {
				//	currentDigest, err = ds.GetCurrentDigest()
				//	logger.Errorf("Failed fetching current digest: %v", err)
				//	break
				//}

				bundle, digest, hasNew := hasNewBundle(ctx, currentDigest, spire)
				if !hasNew {
					break
				}
				logger.Info("Bundle has changed, pushing to Galadriel Server")

				// HA mode push
				//if ds != nil {
				//	harvesterID, err = ds.EnsureID(harvesterID)
				//	if err != nil {
				//		// TODO: handle
				//		logger.Errorf("Failed ensuring harvester ID: %v", err)
				//		break
				//	}
				//	isLeader, err := ds.IsLeader(harvesterID)
				//	if err != nil {
				//		// TODO: handle
				//		logger.Errorf("Failed ensuring harvester ID: %v", err)
				//		break
				//	}
				//	if !isLeader {
				//		break
				//	}
				//}
				req, err := buildPostBundleRequest(bundle)
				if err != nil {
					logger.Error(err)
					break
				}

				//if ds != nil {
				//	// needs to retry
				//	err = ds.UpdateBundle(bundle)
				//	// understand better what to do here (faulty db connection)
				//	if err != nil {
				//		logger.Errorf("Failed to update X.509 bundle: %v", err)
				//		break
				//	}
				//}

				if err = server.PostBundle(ctx, req); err != nil {
					logger.Errorf("Failed to push X.509 bundle: %v", err)
					//if ds != nil {
					//	// blocks until lease TTL
					//	err = ds.RevertBundleUpdate()
					//	if err != nil {
					//		logger.Errorf("Failed reverting bundle update: %v", err)
					//	}
					//}
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

	req := &common.PostBundleRequest{
		TrustBundle: common.TrustBundle{
			TrustDomain:  b.TrustDomain(),
			Bundle:       bundle,
			BundleDigest: util.GetDigest(x509b),
		},
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
		if b.Bundle == nil {
			logger.Errorf("Received an empty bundle for trust domain %q", td)
			continue
		}

		bundle, err := spiffebundle.Parse(td, b.Bundle)
		if err != nil {
			logger.Errorf("Failed to parse trust bundle for %q: %v", td, err)
			continue
		}

		bundles = append(bundles, bundle)
		processed++
	}

	return bundles, processed
}
