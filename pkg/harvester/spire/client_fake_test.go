package spire

import (
	"context"
	"errors"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc"
)

type fakeClientConn struct{}

func (fakeClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}
func (fakeClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
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
	bundle       *spiffebundle.Bundle
	getBundleErr error
}

func (c fakeInternalClient) GetBundle(context.Context) (*spiffebundle.Bundle, error) {
	if c.getBundleErr != nil {
		return nil, c.getBundleErr
	}

	return c.bundle, nil
}

func (c fakeInternalClient) BatchSetFederatedBundle(context.Context, []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error) {
	return nil, errors.New("not implemented")
}

func (c fakeInternalClient) ListFederatedBundles(context.Context) (*ListFederatedBundlesResponse, error) {
	return nil, errors.New("not implemented")
}

func (c fakeInternalClient) GetFederatedBundles(context.Context, []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error) {
	return nil, errors.New("not implemented")
}
