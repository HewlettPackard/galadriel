package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/jmhodges/clock"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/endpoints"
)

// Clock used for time operations that allows to use a Fake for testing
var clk = clock.New()

// Server represents a Galadriel Server.
type Server struct {
	config *Config
	CA     *ca.CA
}

// New creates a new instance of the Galadriel Server.
func New(config *Config) (*Server, error) {
	cert, err := cryptoutil.LoadCertificate(config.CertPath)
	if err != nil {
		return nil, fmt.Errorf("failed loading CA root certificate: %w", err)
	}
	key, err := cryptoutil.LoadRSAPrivateKey(config.CertKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed loading CA root private: %w", err)
	}

	CA, err := ca.New(&ca.Config{
		RootCert: cert,
		RootKey:  key,
		Clock:    clk,
		Logger:   config.Logger.WithField(telemetry.SubsystemName, telemetry.ServerCA),
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating GCA CA: %w", err)
	}

	return &Server{
		config: config,
		CA:     CA,
	}, nil
}

// Run starts running the Galadriel Server, starting its endpoints.
func (s *Server) Run(ctx context.Context) error {
	if err := s.run(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Server) run(ctx context.Context) error {
	endpointsServer, err := s.newEndpointsServer()
	if err != nil {
		return err
	}

	err = util.RunTasks(ctx, endpointsServer.ListenAndServe)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (s *Server) newEndpointsServer() (endpoints.Server, error) {
	config := &endpoints.Config{
		CA:                  s.CA,
		TCPAddress:          s.config.TCPAddress,
		LocalAddress:        s.config.LocalAddress,
		DatastoreConnString: s.config.DBConnString,
		Logger:              s.config.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
	}

	return endpoints.New(config)
}
