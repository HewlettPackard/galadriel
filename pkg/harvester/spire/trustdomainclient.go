package spire

import (
	"context"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	trustdomainv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	"google.golang.org/grpc"
)

type TrustDomainClient interface {
	ListFederationRelationships(context.Context) ([]*FederationRelationship, error)
	CreateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error)
	UpdateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error)
	DeleteFederationRelationships(context.Context, []*spiffeid.TrustDomain) ([]*FederationRelationshipResult, error)
}

type trustDomainClient struct {
	client trustdomainv1.TrustDomainClient
}

func NewTrustDomainClient(cc grpc.ClientConnInterface) TrustDomainClient {
	return trustDomainClient{client: trustdomainv1.NewTrustDomainClient(cc)}
}

func (c trustDomainClient) ListFederationRelationships(ctx context.Context) ([]*FederationRelationship, error) {
	var rels []*FederationRelationship
	var pageToken string
	pageSize := 100

	for {
		res, err := c.client.ListFederationRelationships(ctx, &trustdomainv1.ListFederationRelationshipsRequest{
			PageToken: pageToken,
			PageSize:  int32(pageSize),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list federation relationships: %v", err)
		}
		page, err := protoToFederationsRelationships(res)
		if err != nil {
			return nil, fmt.Errorf("failed to parse federation relationships: %v", err)
		}

		rels = append(rels, page...)

		pageToken = res.NextPageToken
		if res.NextPageToken == "" {
			break
		}

	}

	return rels, nil
}

func (c trustDomainClient) CreateFederationRelationships(ctx context.Context, federationRelationships []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	protoFedRels, err := federationRelationshipsToProto(federationRelationships)
	if err != nil {
		return nil, fmt.Errorf("failed to convert federation relationships to proto: %v", err)
	}

	res, err := c.client.BatchCreateFederationRelationship(ctx, &trustdomainv1.BatchCreateFederationRelationshipRequest{
		FederationRelationships: protoFedRels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create federation relationships: %v", err)
	}

	rels, err := protoCreateToFederationRelatioshipResult(res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse federation relationship results: %v", err)
	}

	return rels, nil
}

func (c trustDomainClient) UpdateFederationRelationships(ctx context.Context, federationRelationships []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	protoFedRels, err := federationRelationshipsToProto(federationRelationships)
	if err != nil {
		return nil, fmt.Errorf("failed to convert federation relationships to proto: %v", err)
	}

	res, err := c.client.BatchUpdateFederationRelationship(ctx, &trustdomainv1.BatchUpdateFederationRelationshipRequest{
		FederationRelationships: protoFedRels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update federation relationships: %v", err)
	}

	rels, err := protoUpdateToFederationRelatioshipResult(res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse federation relationship results: %v", err)
	}

	return rels, nil
}

func (c trustDomainClient) DeleteFederationRelationships(ctx context.Context, trustDomains []*spiffeid.TrustDomain) ([]*FederationRelationshipResult, error) {
	tds, err := trustDomainsToStrings(trustDomains)
	if err != nil {
		return nil, fmt.Errorf("failed to convert trust domains to strings: %v", err)
	}

	res, err := c.client.BatchDeleteFederationRelationship(ctx, &trustdomainv1.BatchDeleteFederationRelationshipRequest{
		TrustDomains: tds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete federation relationships: %v", err)
	}

	rels, err := protoDeleteToFederationRelatioshipResult(res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse federation relationship results: %v", err)
	}

	return rels, nil
}

func trustDomainsToStrings(in []*spiffeid.TrustDomain) ([]string, error) {
	var out []string
	for _, td := range in {
		if td == nil || td.IsZero() {
			return nil, fmt.Errorf("invalid trust domain: %v", td)
		}
		out = append(out, td.String())
	}

	return out, nil
}
