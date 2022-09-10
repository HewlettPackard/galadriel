package spire

import (
	"context"
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SpireServer interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	ListFederationRelationships(context.Context) ([]*FederationRelationship, error)
}

type localSpireServer struct {
	client client
	logger common.Logger
}

type client interface {
	BundleClient
	TrustDomainClient
}

var dialFn = dialSocket
var grpcDialContext = grpc.DialContext

func NewLocalSpireServer(ctx context.Context, socketPath string) SpireServer {
	client, err := dialFn(ctx, socketPath, makeSpireClient)
	if err != nil {
		panic(err)
	}

	return &localSpireServer{
		client: client,
		logger: *common.NewLogger("local_spire_server"),
	}
}

func (s *localSpireServer) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := s.client.GetBundle(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %v", err)
	}

	return bundle, nil
}

func (s *localSpireServer) ListFederationRelationships(ctx context.Context) ([]*FederationRelationship, error) {
	feds, err := s.client.ListFederationRelationships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list federation relationships: %v", err)
	}

	return feds, nil
}

type clientMaker func(*grpc.ClientConn) (client, error)

func dialSocket(ctx context.Context, path string, makeClient clientMaker) (client, error) {
	target, err := common.GetAbsoluteUDSPath(path)
	if err != nil {
		return nil, err
	}

	clientConn, err := grpcDialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial API socket: %v", err)
	}

	client, err := makeClient(clientConn)
	if err != nil {
		return nil, fmt.Errorf("failed to make client: %v", err)
	}

	return client, nil
}

func makeSpireClient(clientConn *grpc.ClientConn) (client, error) {
	if clientConn == nil {
		return nil, errors.New("grpc client connection is invalid")
	}

	return struct {
		BundleClient
		TrustDomainClient
	}{
		BundleClient:      NewBundleClient(clientConn),
		TrustDomainClient: NewTrustDomainClient(clientConn),
	}, nil
}
