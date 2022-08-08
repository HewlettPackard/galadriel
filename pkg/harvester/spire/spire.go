package spire

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/HewlettPackard/Galadriel/pkg/common"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SpireServer interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	ListFederationRelationships(context.Context) ([]*FederationRelationship, error)
	CreateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error)
	UpdateFederationRelationships(context.Context, []*FederationRelationship) ([]*FederationRelationshipResult, error)
	DeleteFederationRelationships(context.Context, []*spiffeid.TrustDomain) ([]*FederationRelationshipResult, error)
}

type localSpireServer struct {
	client client
	logger common.Logger
}

type client interface {
	TrustDomainClient
	BundleClient
}

func NewLocalSpireServer(socketPath string) SpireServer {
	client, err := dialSocket(context.Background(), socketPath)
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

func (s *localSpireServer) CreateFederationRelationships(ctx context.Context, rels []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	res, err := s.client.CreateFederationRelationships(ctx, rels)

	if err != nil {
		return nil, fmt.Errorf("failed to create federation relationships: %v", err)
	}

	if len(res) > len(rels) {
		s.logger.Warn("Creating %d federation relationships returned %d responses", len(rels), len(res))
	}

	return res, nil
}

func (s *localSpireServer) UpdateFederationRelationships(ctx context.Context, rels []*FederationRelationship) ([]*FederationRelationshipResult, error) {
	res, err := s.client.UpdateFederationRelationships(ctx, rels)

	if err != nil {
		return nil, fmt.Errorf("failed to update federation relationships: %v", err)
	}

	if len(res) > len(rels) {
		s.logger.Warn("Updating %d federation relationships returned %d responses", len(rels), len(res))
	}

	return res, nil
}

func (s *localSpireServer) DeleteFederationRelationships(ctx context.Context, trustDomains []*spiffeid.TrustDomain) ([]*FederationRelationshipResult, error) {
	res, err := s.client.DeleteFederationRelationships(ctx, trustDomains)
	if err != nil {
		return nil, fmt.Errorf("failed to delete federarion relationships: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to delete federation relationships: %v", err)
	}

	if len(res) > len(trustDomains) {
		s.logger.Warn("Deleting %d federation relationships returned %d responses", len(trustDomains), len(res))
	}

	return res, nil

}

func dialSocket(ctx context.Context, path string) (client, error) {
	var target string

	if filepath.IsAbs(path) {
		target = "unix://" + path
	} else {
		target = "unix:" + path
	}
	grpcClient, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial API socket: %v", err)
	}

	return struct {
		TrustDomainClient
		BundleClient
		io.Closer
	}{
		TrustDomainClient: NewTrustDomainClient(grpcClient),
		BundleClient:      NewBundleClient(grpcClient),
		Closer:            grpcClient,
	}, nil
}
