package spire

import (
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/grpc/codes"
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
	Status                 *FederationRelationshipResultStatus
	FederationRelationship *FederationRelationship
}

type FederationRelationshipResultStatus struct {
	Code        codes.Code
	Message     string
	TrustDomain string
}
