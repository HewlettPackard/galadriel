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
}

func NewBundleClient(cc grpc.ClientConnInterface) BundleClient {
	return bundleClient{client: bundlev1.NewBundleClient(cc)}
}

type bundleClient struct {
	client bundlev1.BundleClient
}

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
