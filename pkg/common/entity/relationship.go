package entity

import (
	"net/url"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire/proto/spire/common"
)

type BundleEndpointType string

type Relationship struct {
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     *url.URL
	BundleEndpointProfile BundleEndpointType
	TrustDomainBundle     *common.Bundle

	// Fields only used for 'https_spiffe' bundle endpoint profile
	EndpointSPIFFEID spiffeid.ID
}
