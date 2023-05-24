package models

import (
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// BundlesDigests is a map of trust bundle digests keyed by trust domain.
type BundlesDigests map[spiffeid.TrustDomain][]byte

// BundleUpdates is a map of trust bundles keyed by trust domain.
type BundleUpdates map[spiffeid.TrustDomain]*entity.Bundle

// SyncBundleRequest represents a request to send the current state of federated bundles digests.
type SyncBundleRequest struct {
	State BundlesDigests `json:"state"`
}

// SyncBundleResponse represents a response from Galadriel Server containing the
// federated trust bundles updates.
type SyncBundleResponse struct {
	// Update conveys trust bundles that are new or updates.
	Updates BundleUpdates `json:"updates"`

	// State is the current source-of-truth map of all trust bundles.
	// It essentially allows triggering deletions of trust bundles on harvesters.
	State BundlesDigests `json:"state"`
}

// PostBundleRequest represents the request to submit the local SPIRE Server's bundle.
type PostBundleRequest struct {
	*entity.Bundle `json:"state"`
}

// ConvertSPIFFEBundleToEntityBundle converts a SPIFFE bundle to a Galadriel bundle.
func ConvertSPIFFEBundleToEntityBundle(spiffeBundle *spiffebundle.Bundle) (*entity.Bundle, error) {
	bytes, err := spiffeBundle.Marshal()
	if err != nil {
		return nil, err
	}

	bundle := &entity.Bundle{
		TrustDomainName: spiffeBundle.TrustDomain(),
		Data:            bytes,
		Digest:          cryptoutil.CalculateDigest(bytes),
	}

	return bundle, nil
}

// ConvertEntityBundleToSPIFFEBundle converts a Galadriel bundle to a SPIFFE bundle.
func ConvertEntityBundleToSPIFFEBundle(galadrielBundle *entity.Bundle) (*spiffebundle.Bundle, error) {
	bundle, err := spiffebundle.Parse(galadrielBundle.TrustDomainName, galadrielBundle.Data)
	if err != nil {
		return nil, err
	}

	return bundle, nil
}
