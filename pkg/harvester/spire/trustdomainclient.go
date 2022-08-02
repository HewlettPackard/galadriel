package spire

import (
	"context"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	trustdomainv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	apitypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc"
)

type FederationRelationship struct {
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     string
	BundleEndpointProfile BundleEndpointProfile
	TrustDomainBundle     *spiffebundle.Bundle
}

type TrustDomainClient interface {
	ListFederationRelationships(context.Context) ([]FederationRelationship, error)
	// CreateFederationRelationships(ctx context.Context, federationRelationships []FederationRelationship) ([]Status, error)
	// UpdateFederationRelationships(ctx context.Context, federationRelationships []FederationRelationship) ([]Status, error)
	// DeleteFederationRelationships(ctx context.Context, tds []spiffeid.TrustDomain) ([]Status, error)
}

type trustDomainClient struct {
	client trustdomainv1.TrustDomainClient
}

func NewTrustDomainClient(cc grpc.ClientConnInterface) TrustDomainClient {
	return trustDomainClient{client: trustdomainv1.NewTrustDomainClient(cc)}
}

func (c trustDomainClient) ListFederationRelationships(ctx context.Context) ([]FederationRelationship, error) {
	res, err := c.client.ListFederationRelationships(ctx, &trustdomainv1.ListFederationRelationshipsRequest{})
	if err != nil {
		return nil, err
	}
	rels, err := parseFederationsRelationships(res)
	if err != nil {
		return nil, err
	}

	return rels, nil
}

func parseFederationsRelationships(in *trustdomainv1.ListFederationRelationshipsResponse) ([]FederationRelationship, error) {
	var out []FederationRelationship
	for _, inRel := range in.FederationRelationships {
		td, err := spiffeid.TrustDomainFromString(inRel.TrustDomain)
		if err != nil {
			return nil, fmt.Errorf("failed parsing federated trust domain: %v", err)
		}
		bundle, err := parseBundle(inRel.TrustDomainBundle)
		if err != nil {
			return nil, fmt.Errorf("failed parsing federated trust bundle: %v", err)
		}
		profile, err := parseBundleProfile(inRel)
		if err != nil {
			return nil, fmt.Errorf("failed parsing federated profile: %v", err)
		}

		outRel := FederationRelationship{
			TrustDomain:           td,
			TrustDomainBundle:     bundle,
			BundleEndpointURL:     inRel.BundleEndpointUrl,
			BundleEndpointProfile: profile,
		}
		out = append(out, outRel)
	}

	return out, nil
}

func parseBundleProfile(in *apitypes.FederationRelationship) (BundleEndpointProfile, error) {
	var out BundleEndpointProfile
	switch in.BundleEndpointProfile.(type) {
	case *apitypes.FederationRelationship_HttpsWeb:
		out = HTTPSWebBundleEndpointProfile{}
	case *apitypes.FederationRelationship_HttpsSpiffe:
		spiffeId, err := spiffeid.FromString(in.GetHttpsSpiffe().EndpointSpiffeId)
		if err != nil {
			return nil, err
		}
		out = HTTPSSpiffeBundleEndpointProfile{
			SpiffeId: spiffeId,
		}
	}

	return out, nil
}

type BundleEndpointProfile interface{}

type HTTPSWebBundleEndpointProfile struct{}

type HTTPSSpiffeBundleEndpointProfile struct {
	SpiffeId spiffeid.ID
}
