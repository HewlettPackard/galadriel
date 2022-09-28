package common

// FederationState maps MemberState's by their trust domain.
type FederationState map[string]MemberState

// MemberState defines a member to be updated.
// All returned MemberState's implicitly federates with the calling Harvester.
type MemberState struct {
	TrustDomain     string `json:"trust_domain"`
	TrustBundle     []byte `json:"trust_bundle"`
	TrustBundleHash []byte `json:"trust_bundle_hash"`
}

// SyncBundleBody is the request body for the /bundle/sync endpoint.
type SyncBundleBody struct {
	// FederatesWith maps (keyed by trust-domain) which members federate with the calling harvester.
	// In other words, entry in this map represents a relationship between the calling harvester and Member.
	// This is so the server can know what the calling harvester needs to be updated of.
	FederatesWith FederationState `json:"federates_with"`
}

// SyncBundleResponse is the request response for the /bundle/sync endpoint.
type SyncBundleResponse struct {
	// Update represent bundle sets that need to be performed by the calling harvester.
	Update FederationState `json:"update"`
	// State is the current source-of-truth list of relationships.
	// It essentially allows triggering deletions on calling harvesters.
	State FederationState `json:"state"`
}

// PostBundleBody is the request body for the /bundle endpoint.
type PostBundleBody struct {
	// MemberState is the latest watched state for a current member (SPIRE Server)
	MemberState
}

// PostBundleResponse is the request response for the /bundle endpoint
type PostBundleResponse struct{}
