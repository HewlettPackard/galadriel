package spire

import (
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type FederationRelationship struct {
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     string
	BundleEndpointProfile BundleEndpointProfile
	TrustDomainBundle     *spiffebundle.Bundle
}

type BundleEndpointProfile interface {
	Name() string
}

type HTTPSWebBundleEndpointProfile struct{}

func (HTTPSWebBundleEndpointProfile) Name() string {
	return "https_web"
}

type HTTPSSpiffeBundleEndpointProfile struct {
	SpiffeID spiffeid.ID
}

func (HTTPSSpiffeBundleEndpointProfile) Name() string {
	return "https_spiffe"
}

type FederationRelationshipResult struct {
	status                 *FederationRelationshipResultStatus
	federationRelationship *FederationRelationship
}

type FederationRelationshipResultStatus struct {
	// A status code, which should be an enum value of google.rpc.Code.
	code        int32
	message     string
	trustDomain string
}
