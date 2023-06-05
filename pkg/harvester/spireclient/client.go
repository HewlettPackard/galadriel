package spireclient

import (
	"context"
	"fmt"
	"net"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	listFederatedBundlesPageSize = 100
	defaultSocketPath            = "/tmp/spire-server/private/api.sock"
)

// Client is an interface for interacting with a SPIRE Server, providing methods for trust bundle retrieval,
// setting federation bundles, and deleting federation bundles.
type Client interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	GetFederatedBundles(context.Context) ([]*spiffebundle.Bundle, error)
	SetFederatedBundles(context.Context, []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error)
	DeleteFederatedBundles(context.Context, []spiffeid.TrustDomain) ([]*BatchDeleteFederatedBundleStatus, error)
}

type spireServerClient struct {
	bundleClient bundlev1.BundleClient
}

func NewSpireClient(ctx context.Context, addr net.Addr) (Client, error) {
	if addr == nil {
		addr = &net.UnixAddr{
			Name: defaultSocketPath,
			Net:  "unix",
		}
	}
	clientConn, err := dialSocket(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &spireServerClient{
		bundleClient: bundlev1.NewBundleClient(clientConn),
	}, nil
}

func (c *spireServerClient) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := c.bundleClient.GetBundle(ctx, &bundlev1.GetBundleRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %v", err)
	}

	spiffeBundle, err := protoToBundle(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spire server bundle response: %v", err)
	}

	return spiffeBundle, nil
}

// GetFederatedBundles lists all the SPIFFE bundles the SPIRE Server set in SPIRE.
func (c *spireServerClient) GetFederatedBundles(ctx context.Context) ([]*spiffebundle.Bundle, error) {
	var pageToken string
	result := make([]*spiffebundle.Bundle, 0)

	for {
		res, err := c.bundleClient.ListFederatedBundles(ctx, &bundlev1.ListFederatedBundlesRequest{
			PageToken: pageToken,
			PageSize:  int32(listFederatedBundlesPageSize),
		})
		if err != nil {
			return nil, err
		}

		bundles, err := protoToSpiffeBundles(res)
		if err != nil {
			return nil, err
		}

		result = append(result, bundles...)

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	return result, nil
}

func (c *spireServerClient) DeleteFederatedBundles(ctx context.Context, trustDomains []spiffeid.TrustDomain) ([]*BatchDeleteFederatedBundleStatus, error) {
	tdsToDel := make([]string, 0, len(trustDomains))
	for _, td := range trustDomains {
		tdsToDel = append(tdsToDel, td.String())
	}

	resp, err := c.bundleClient.BatchDeleteFederatedBundle(ctx, &bundlev1.BatchDeleteFederatedBundleRequest{
		TrustDomains: tdsToDel,
		Mode:         bundlev1.BatchDeleteFederatedBundleRequest_DISSOCIATE, // we don't want to delete entries, just dissociate them from the deleted bundle
	})
	if err != nil {
		return nil, fmt.Errorf("client failed to delete federated bundles: %v", err)
	}

	statuses, err := protoToBatchDeleteFederatedBundleResult(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spire server bundle response: %v", err)
	}

	return statuses, err
}

// SetFederatedBundles adds or updates a set of federated SPIFFE bundles on the SPIRE Server
func (c *spireServerClient) SetFederatedBundles(ctx context.Context, bundles []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error) {
	protoBundles, err := bundlesToProto(bundles)
	if err != nil {
		return nil, err
	}

	resp, err := c.bundleClient.BatchSetFederatedBundle(ctx, &bundlev1.BatchSetFederatedBundleRequest{Bundle: protoBundles})
	if err != nil {
		return nil, fmt.Errorf("client failed to set federated bundles: %v", err)
	}

	statuses, err := protoToBatchSetFederatedBundleResult(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spire server bundle response: %v", err)
	}

	return statuses, nil
}

func dialSocket(ctx context.Context, addr net.Addr) (*grpc.ClientConn, error) {
	target := fmt.Sprintf("%s://%s", addr.Network(), addr.String())
	clientConn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial API socket: %v", err)
	}

	return clientConn, nil
}

func newLocalSpireServerWithClient(bundleClient bundlev1.BundleClient) Client {
	return &spireServerClient{
		bundleClient: bundleClient,
	}
}
