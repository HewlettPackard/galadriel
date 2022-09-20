package client

import (
	"context"
	"errors"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
)

// GaladrielServerClient represents a client to connect to Galadriel Server
type GaladrielServerClient interface {
	GetUpdates(context.Context) ([]string, error)
	PushUpdates(context.Context, []string) error
}

type client struct {
	address string
	logger  logrus.FieldLogger
}

func NewGaladrielServerClient(address string) (GaladrielServerClient, error) {
	return &client{
		address: address,
		logger:  logrus.WithField(telemetry.SubsystemName, telemetry.GaladrielServerClient),
	}, nil
}

func (s *client) GetUpdates(ctx context.Context) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (s *client) PushUpdates(ctx context.Context, updates []string) error {
	return errors.New("not implemented")
}
