package common

// FederationState maps MemberState's by their trust domain.
type FederationState map[string]MemberState

// MemberState is the minimum necessary to define a member state.
// All returned MemberState's implicitly federate with the calling Harvester.
type MemberState struct {
	TrustDomain     string `json:"trust_domain"`
	TrustBundle     []byte `json:"trust_bundle"`
	TrustBundleHash []byte
}

// SyncBundleRequest is the request body for the /bundle/sync endpoint.
type SyncBundleRequest struct {
	// FederatesWith maps to (keyed by trust-domain) which members federate with the calling harvester.
	// In other words, each entry in this map represents a relationship between the calling harvester the MemberState.
	// With this, the server can know what the calling harvester needs to be updated of.
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

// PostBundleRequest is the request body for the /bundle endpoint.
type PostBundleRequest struct {
	// MemberState is the latest watched state from the calling harvester.
	MemberState `json:"state"`
}
