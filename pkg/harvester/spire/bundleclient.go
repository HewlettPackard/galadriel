package spire

import (
	"context"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	"google.golang.org/grpc"
)

type BundleClient interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	BatchSetFederatedBundle(context.Context, []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error)
}

// NewBundleClient creates a new SPIRE Bundle API client
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
