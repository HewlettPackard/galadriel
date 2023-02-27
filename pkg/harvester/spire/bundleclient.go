package spire

import (
	"context"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	"google.golang.org/grpc"
)

const listFederatedBundlesPageSize = 100

type BundleClient interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	BatchSetFederatedBundle(context.Context, []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error)
	ListFederatedBundles(context.Context) (*ListFederatedBundlesResponse, error)
}

// NewBundleClient creates a new SPIRE Data API client
func NewBundleClient(cc grpc.ClientConnInterface) BundleClient {
	return bundleClient{client: bundlev1.NewBundleClient(cc)}
}

type bundleClient struct {
	client bundlev1.BundleClient
}

// GetBundle returns the current bundle of the server
func (c bundleClient) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := c.client.GetBundle(ctx, &bundlev1.GetBundleRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle from trust domain client: %v", err)
	}

	spiffeBundle, err := protoToBundle(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spire server bundle response: %v", err)
	}

	return spiffeBundle, nil
}

// ListFederatedBunles retrieves all the bundles the server knows about
func (c bundleClient) ListFederatedBundles(ctx context.Context) (*ListFederatedBundlesResponse, error) {
	var out ListFederatedBundlesResponse
	var pageToken string

	for {
		res, err := c.client.ListFederatedBundles(ctx, &bundlev1.ListFederatedBundlesRequest{
			PageToken: pageToken,
			PageSize:  int32(listFederatedBundlesPageSize),
		})
		if err != nil {
			return nil, err
		}

		bundles, err := protoToFederatedBundles(res)
		if err != nil {
			return nil, err
		}

		out.Bundles = append(out.Bundles, bundles...)

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	return &out, nil
}

// BatchSetFederatedBundle adds or updates federated bundles
func (c bundleClient) BatchSetFederatedBundle(ctx context.Context, bundles []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error) {
	protoBundles, err := bundlesToProto(bundles)
	if err != nil {
		return nil, err
	}

	res, err := c.client.BatchSetFederatedBundle(ctx, &bundlev1.BatchSetFederatedBundleRequest{Bundle: protoBundles})
	if err != nil {
		return nil, fmt.Errorf("client failed to set federated bundles: %v", err)
	}

	statuses, err := protoToBatchSetFederatedBundleResult(res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spire server bundle response: %v", err)
	}

	return statuses, nil
}
