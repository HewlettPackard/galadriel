package server

import (
	"context"
	"errors"

	"github.com/HewlettPackard/Galadriel/pkg/common"
)

type GaladrielServer interface {
	GetUpdates(context.Context) ([]string, error)
	PushUpdates(context.Context, []string) error
	GetMemberships(context.Context) ([]string, error)
}

type RemoteGaladrielServer struct {
	address string
	logger  common.Logger
}

func NewRemoteGaladrielServer(address string) GaladrielServer {
	return &RemoteGaladrielServer{
		address: address,
		logger:  *common.NewLogger("remote_galadriel_server"),
	}
}

func (s *RemoteGaladrielServer) GetUpdates(ctx context.Context) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (s *RemoteGaladrielServer) PushUpdates(ctx context.Context, updates []string) error {
	return errors.New("not implemented")
}

func (s *RemoteGaladrielServer) GetMemberships(ctx context.Context) ([]string, error) {
	return nil, errors.New("not implemented")
}
