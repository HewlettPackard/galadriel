package common

type PostBundleRequest struct {
	TrustDomain string `json:"trust_domain"`
	TrustBundle []byte `json:"trust_bundle"`
}
