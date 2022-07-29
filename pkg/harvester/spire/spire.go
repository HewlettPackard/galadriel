package spire

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/HewlettPackard/Galadriel/pkg/common"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/spire-controller-manager/pkg/spireapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SpireServer interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	GetFederationRelationships(context.Context) ([]spireapi.FederationRelationship, error)
	CreateFederationRelationship(context.Context, *spiffebundle.Bundle) (*spireapi.Status, error)
}

type LocalSpireServer struct {
	client Client
	logger common.Logger
}

func NewLocalSpireServer(socketPath string) SpireServer {
	client, err := dialSocket(context.Background(), socketPath)
	if err != nil {
		panic(err)
	}
	return &LocalSpireServer{
		client: client,
		logger: *common.NewLogger("local_spire_server"),
	}
}

func (s *LocalSpireServer) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := s.client.GetBundle(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	return bundle, nil
}

func (s *LocalSpireServer) GetFederationRelationships(ctx context.Context) ([]spireapi.FederationRelationship, error) {
	feds, err := s.client.ListFederationRelationships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list federation relationships: %w", err)
	}
	return feds, nil
}

func (s *LocalSpireServer) CreateFederationRelationship(ctx context.Context, bundle *spiffebundle.Bundle) (*spireapi.Status, error) {
	x509bundle := bundle.X509Bundle()

	s.logger.Debug("Creating federation relationship with", bundle.TrustDomain().ID())
	spireSpiffeId, _ := bundle.TrustDomain().ID().AppendPath("/spire/server")

	status, err := s.client.CreateFederationRelationships(ctx, []spireapi.FederationRelationship{
		{
			TrustDomain:       x509bundle.TrustDomain(),
			TrustDomainBundle: bundle,
			// TODO: pass this in as a parameter
			BundleEndpointURL: "https://localhost:8442",
			// TODO: pass this in as a parameter
			BundleEndpointProfile: spireapi.HTTPSSPIFFEProfile{
				EndpointSPIFFEID: spireSpiffeId,
			},
		},
	})

	if err != nil || len(status) == 0 {
		return nil, fmt.Errorf("failed to create federation relationship: %w", err)
	}

	if len(status) > 1 {
		s.logger.Warn("Creating a single federation relationship returned multiple responses", status)
	}

	return &status[0], nil
}

func dialSocket(ctx context.Context, path string) (Client, error) {
	var target string

	if filepath.IsAbs(path) {
		target = "unix://" + path
	} else {
		target = "unix:" + path
	}
	grpcClient, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial API socket: %w", err)
	}

	return struct {
		spireapi.TrustDomainClient
		spireapi.BundleClient
		io.Closer
	}{
		TrustDomainClient: spireapi.NewTrustDomainClient(grpcClient),
		BundleClient:      spireapi.NewBundleClient(grpcClient),
		Closer:            grpcClient,
	}, nil
}

type Client interface {
	spireapi.TrustDomainClient
	spireapi.BundleClient
}
