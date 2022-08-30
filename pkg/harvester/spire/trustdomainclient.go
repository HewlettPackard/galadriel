package spire

import (
	"context"
	"fmt"

	trustdomainv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	"google.golang.org/grpc"
)

var listFederationRelationshipsPageSize = 100

type TrustDomainClient interface {
	ListFederationRelationships(context.Context) ([]*FederationRelationship, error)
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

	for {
		res, err := c.client.ListFederationRelationships(ctx, &trustdomainv1.ListFederationRelationshipsRequest{
			PageToken: pageToken,
			PageSize:  int32(listFederationRelationshipsPageSize),
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
