package spire

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: can we change the name to SpireServerLocalClient?
type SpireServer interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	SetFederatedBundles(context.Context, []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error)
}

type localSpireServer struct {
	client client
	log    logrus.FieldLogger
}

type client interface {
	BundleClient
}

var dialFn = dialSocket
var grpcDialContext = grpc.DialContext

// NewLocalSpireServer creates and initializes a new client to communicate with
// a SPIRE Server given an address to its admin API
func NewLocalSpireServer(ctx context.Context, addr net.Addr) SpireServer {
	client, err := dialFn(ctx, addr, makeSpireClient)
	if err != nil {
		panic(err)
	}

	return &localSpireServer{
		client: client,
		log:    logrus.WithField(telemetry.SubsystemName, "local_spire_server"),
	}
}

// GetBundle returns the current Trust Bundle of the SPIRE Server
func (s *localSpireServer) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := s.client.GetBundle(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}

	return bundle, nil
}

// SetFederatedBundles adds or updates a set of federated bundles on a SPIRE Server
func (s *localSpireServer) SetFederatedBundles(ctx context.Context, bundles []*spiffebundle.Bundle) ([]*BatchSetFederatedBundleStatus, error) {
	res, err := s.client.BatchSetFederatedBundle(ctx, bundles)
	if err != nil {
		return nil, fmt.Errorf("failed to set federated bundles: %v", err)
	}

	return res, nil
}

type clientMaker func(*grpc.ClientConn) (client, error)

func dialSocket(ctx context.Context, addr net.Addr, makeClient clientMaker) (client, error) {
	target := fmt.Sprintf("%s://%s", addr.Network(), addr.String())
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
	}{
		BundleClient: NewBundleClient(clientConn),
	}, nil
}
