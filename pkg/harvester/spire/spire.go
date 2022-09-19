package spire

import (
	"context"
	"errors"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
	"path/filepath"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: can we change the name to SpireServerLocalClient?
type SpireServer interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
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

func NewLocalSpireServer(ctx context.Context, socketPath string) SpireServer {
	client, err := dialFn(ctx, socketPath, makeSpireClient)
	if err != nil {
		panic(err)
	}

	return &localSpireServer{
		client: client,
		log:    logrus.WithField(telemetry.SubsystemName, "local_spire_server"),
	}
}

func (s *localSpireServer) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := s.client.GetBundle(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}

	return bundle, nil
}

type clientMaker func(*grpc.ClientConn) (client, error)

func dialSocket(ctx context.Context, path string, makeClient clientMaker) (client, error) {
	var target string

	if filepath.IsAbs(path) {
		target = "unix://" + path
	} else {
		target = "unix:" + path
	}
	clientConn, err := grpcDialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial API socket: %w", err)
	}

	client, err := makeClient(clientConn)
	if err != nil {
		return nil, fmt.Errorf("failed to make client: %w", err)
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
