package spire

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	trustdomainv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakeClientConn struct{}

func (fakeClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return errors.New("not implemented")
}
func (fakeClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}

type fakeSpireBundleClient struct {
	bundle       *types.Bundle
	getBundleErr error
}

func (c fakeSpireBundleClient) GetBundle(ctx context.Context, in *bundlev1.GetBundleRequest, opts ...grpc.CallOption) (*types.Bundle, error) {
	if c.getBundleErr != nil {
		return nil, c.getBundleErr
	}

	return c.bundle, nil
}

func (c fakeSpireBundleClient) CountBundles(ctx context.Context, in *bundlev1.CountBundlesRequest, opts ...grpc.CallOption) (*bundlev1.CountBundlesResponse, error) {
	return nil, errors.New("not implemented")
}

func (fc fakeSpireBundleClient) AppendBundle(ctx context.Context, in *bundlev1.AppendBundleRequest, opts ...grpc.CallOption) (*types.Bundle, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) PublishJWTAuthority(ctx context.Context, in *bundlev1.PublishJWTAuthorityRequest, opts ...grpc.CallOption) (*bundlev1.PublishJWTAuthorityResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) ListFederatedBundles(ctx context.Context, in *bundlev1.ListFederatedBundlesRequest, opts ...grpc.CallOption) (*bundlev1.ListFederatedBundlesResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) GetFederatedBundle(ctx context.Context, in *bundlev1.GetFederatedBundleRequest, opts ...grpc.CallOption) (*types.Bundle, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) BatchCreateFederatedBundle(ctx context.Context, in *bundlev1.BatchCreateFederatedBundleRequest, opts ...grpc.CallOption) (*bundlev1.BatchCreateFederatedBundleResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) BatchUpdateFederatedBundle(ctx context.Context, in *bundlev1.BatchUpdateFederatedBundleRequest, opts ...grpc.CallOption) (*bundlev1.BatchUpdateFederatedBundleResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) BatchSetFederatedBundle(ctx context.Context, in *bundlev1.BatchSetFederatedBundleRequest, opts ...grpc.CallOption) (*bundlev1.BatchSetFederatedBundleResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireBundleClient) BatchDeleteFederatedBundle(ctx context.Context, in *bundlev1.BatchDeleteFederatedBundleRequest, opts ...grpc.CallOption) (*bundlev1.BatchDeleteFederatedBundleResponse, error) {
	return nil, errors.New("not implemented")
}

type fakeInternalClient struct {
	bundle                        *spiffebundle.Bundle
	getBundleErr                  error
	federationRelationships       []*FederationRelationship
	getFederationRelationshipsErr error
}

func (c fakeInternalClient) GetBundle(context.Context) (*spiffebundle.Bundle, error) {
	if c.getBundleErr != nil {
		return nil, c.getBundleErr
	}

	return c.bundle, nil
}

func (c fakeInternalClient) ListFederationRelationships(context.Context) ([]*FederationRelationship, error) {
	if c.getFederationRelationshipsErr != nil {
		return nil, c.getFederationRelationshipsErr
	}

	return c.federationRelationships, nil
}

func (c fakeInternalClient) CreateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	return nil, errors.New("not implemented")
}

func (c fakeInternalClient) UpdateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	return nil, errors.New("not implemented")
}

func (c fakeInternalClient) DeleteFederationRelationships(context.Context, []*spiffeid.TrustDomain) ([]*FederationRelationshipResult, error) {
	return nil, errors.New("not implemented")
}

type fakeSpireTrustDomainClient struct {
	federationRelationships               []*types.FederationRelationship
	batchListFederationRelationshipsError error
}

func (c fakeSpireTrustDomainClient) ListFederationRelationships(ctx context.Context, in *trustdomainv1.ListFederationRelationshipsRequest, opts ...grpc.CallOption) (*trustdomainv1.ListFederationRelationshipsResponse, error) {
	if c.batchListFederationRelationshipsError != nil {
		return nil, c.batchListFederationRelationshipsError
	}

	var start int
	var end int
	var pageToken string

	if in.PageToken == "" {
		start = 0
	} else {
		s, err := strconv.Atoi(in.PageToken)
		if err != nil {
			return nil, fmt.Errorf("invalid page token: %s", in.PageToken)
		}
		start = s - 1
	}
	end = start + int(in.PageSize)

	if end > len(c.federationRelationships) {
		end = len(c.federationRelationships)
	}
	if end < len(c.federationRelationships) && end > 0 {
		pageToken = fmt.Sprint(end + 1)
	}

	out := &trustdomainv1.ListFederationRelationshipsResponse{
		FederationRelationships: c.federationRelationships[start:end],
		NextPageToken:           pageToken,
	}

	return out, nil
}

func (c fakeSpireTrustDomainClient) GetFederationRelationship(ctx context.Context, in *trustdomainv1.GetFederationRelationshipRequest, opts ...grpc.CallOption) (*types.FederationRelationship, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireTrustDomainClient) BatchCreateFederationRelationship(ctx context.Context, in *trustdomainv1.BatchCreateFederationRelationshipRequest, opts ...grpc.CallOption) (*trustdomainv1.BatchCreateFederationRelationshipResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireTrustDomainClient) BatchUpdateFederationRelationship(ctx context.Context, in *trustdomainv1.BatchUpdateFederationRelationshipRequest, opts ...grpc.CallOption) (*trustdomainv1.BatchUpdateFederationRelationshipResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireTrustDomainClient) BatchDeleteFederationRelationship(ctx context.Context, in *trustdomainv1.BatchDeleteFederationRelationshipRequest, opts ...grpc.CallOption) (*trustdomainv1.BatchDeleteFederationRelationshipResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeSpireTrustDomainClient) RefreshBundle(ctx context.Context, in *trustdomainv1.RefreshBundleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, errors.New("not implemented")
}
