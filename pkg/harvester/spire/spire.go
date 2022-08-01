package spire

import (
	"context"
	"errors"

	"github.com/HewlettPackard/Galadriel/pkg/common"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/spire-controller-manager/pkg/spireapi"
)

type SpireServer interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
	GetFederationRelationships(context.Context) ([]spireapi.FederationRelationship, error)
	CreateFederationRelationship(context.Context, *spiffebundle.Bundle) (*spireapi.Status, error)
}

type LocalSpireServer struct {
	logger common.Logger
}

func NewLocalSpireServer(socketPath string) SpireServer {
	return &LocalSpireServer{
		logger: *common.NewLogger("local_spire_server"),
	}
}

func (s *LocalSpireServer) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	return nil, errors.New("not implemented")
}

func (s *LocalSpireServer) GetFederationRelationships(ctx context.Context) ([]spireapi.FederationRelationship, error) {
	return nil, errors.New("not implemented")
}

func (s *LocalSpireServer) CreateFederationRelationship(ctx context.Context, bundle *spiffebundle.Bundle) (*spireapi.Status, error) {
	return nil, errors.New("not implemented")
}
