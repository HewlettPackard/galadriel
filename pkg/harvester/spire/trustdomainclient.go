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

type BundleEndpointProfile interface{}

type HTTPSWebBundleEndpointProfile struct{}

type HTTPSSpiffeBundleEndpointProfile struct {
	SpiffeID spiffeid.ID
}

type FederationRelationshipResult struct {
	status                 *Status
	federationRelationship *FederationRelationship
}

type Status struct {
	// A status code, which should be an enum value of google.rpc.Code.
	code    int32
	message string
}

type TrustDomainClient interface {
	ListFederationRelationships(context.Context) ([]*FederationRelationship, error)
	CreateFederationReslationships(ctx context.Context, federationRelationships []FederationRelationship) ([]*FederationRelationshipResult, error)
	// UpdateFederationRelationships(ctx context.Context, federationRelationships []FederationRelationship) ([]Status, error)
	// DeleteFederationRelationships(ctx context.Context, tds []spiffeid.TrustDomain) ([]Status, error)
}

type trustDomainClient struct {
	client trustdomainv1.TrustDomainClient
}

func NewTrustDomainClient(cc grpc.ClientConnInterface) TrustDomainClient {
	return trustDomainClient{client: trustdomainv1.NewTrustDomainClient(cc)}
}

func (c trustDomainClient) ListFederationRelationships(ctx context.Context) ([]*FederationRelationship, error) {
	res, err := c.client.ListFederationRelationships(ctx, &trustdomainv1.ListFederationRelationshipsRequest{})
	if err != nil {
		return nil, err
	}
	rels, err := protoToFederationsRelationships(res)
	if err != nil {
		return nil, err
	}

	return rels, nil
}

func (c trustDomainClient) CreateFederationReslationships(ctx context.Context, federationRelationships []FederationRelationship) ([]*FederationRelationshipResult, error) {
	res, err := c.client.BatchCreateFederationRelationship(ctx, &trustdomainv1.BatchCreateFederationRelationshipRequest{
		FederationRelationships: federationRelationshipsToProto(federationRelationships),
	})
	if err != nil {
		return nil, err
	}
	rels, err := parseResults(res)
	if err != nil {
		return nil, err
	}

	return rels, nil
}

func parseResults(in *trustdomainv1.BatchCreateFederationRelationshipResponse) ([]*FederationRelationshipResult, error) {
	out := make([]*FederationRelationshipResult, len(in.Results))

	for _, r := range in.Results {
		frel, err := protoToFederationsRelationship(r.FederationRelationship)
		if err != nil {
			return nil, fmt.Errorf("failed to convert federation relationship to proto: %v", err)
		}
		rOut := &FederationRelationshipResult{
			status: &Status{
				code:    r.Status.Code,
				message: r.Status.Message,
			},
			federationRelationship: frel,
		}
		out = append(out, rOut)
	}

	return out, nil
}

func protoToFederationsRelationships(in *trustdomainv1.ListFederationRelationshipsResponse) ([]*FederationRelationship, error) {
	var out []*FederationRelationship
	for _, inRel := range in.FederationRelationships {
		outRel, err := protoToFederationsRelationship(inRel)
		if err != nil {
			return nil, fmt.Errorf("failed parsing federated relationship: %v", err)
		}
		out = append(out, outRel)
	}

	return out, nil
}

func protoToFederationsRelationship(in *apitypes.FederationRelationship) (*FederationRelationship, error) {
	td, err := spiffeid.TrustDomainFromString(in.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("failed parsing federated trust domain: %v", err)
	}
	bundle, err := protoToBundle(in.TrustDomainBundle)
	if err != nil {
		return nil, fmt.Errorf("failed parsing federated trust bundle: %v", err)
	}
	profile, err := protoToBundleProfile(in)
	if err != nil {
		return nil, fmt.Errorf("failed parsing federated profile: %v", err)
	}

	out := &FederationRelationship{
		TrustDomain:           td,
		TrustDomainBundle:     bundle,
		BundleEndpointURL:     in.BundleEndpointUrl,
		BundleEndpointProfile: profile,
	}

	return out, nil
}

func protoToBundleProfile(in *apitypes.FederationRelationship) (BundleEndpointProfile, error) {
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
			SpiffeID: spiffeId,
		}
	}

	return out, nil
}
