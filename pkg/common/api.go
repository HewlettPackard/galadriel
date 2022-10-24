package common

import "github.com/spiffe/go-spiffe/v2/spiffeid"

// BundlesDigests is a map of trust bundle digests keyed by trust domain.
type BundlesDigests map[spiffeid.TrustDomain][]byte

// BundleUpdates is a map of trust bundles keyed by trust domain.
type BundleUpdates map[spiffeid.TrustDomain]TrustBundle

// TrustBundle represents a SPIFFE Trust bundle along with its digest.
type TrustBundle struct {
	// Trust Domain of the bundle
	TrustDomain spiffeid.TrustDomain `json:"trust_domain"`
	// SPIFFE bundle according to the SPIFFE Trust Domain and Bundle specification.
	// https://github.com/spiffe/spiffe/blob/main/standards/SPIFFE_Trust_Domain_and_Bundle.md#4-spiffe-bundle-format
	Bundle []byte `json:"trust_bundle"`
	// SHA3-256 digest of the PEM-encoded X.509 bundle certificate blocks.
	BundleDigest []byte `json:"bundle_digest"`
}

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
	// TrustBundle is the latest watched SPIRE Server trust bundle.
	TrustBundle `json:"state"`
}
