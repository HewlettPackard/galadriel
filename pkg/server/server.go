package server

import (
	"context"
	"errors"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/HewlettPackard/galadriel/pkg/server/endpoints"
)

// Server represents a Galadriel Server.
type Server struct {
	config *Config
}

// New creates a new instance of the Galadriel Server.
func New(config *Config) *Server {
	return &Server{config: config}
}

// Run starts running the Galadriel Server, starting its endpoints.
func (s *Server) Run(ctx context.Context) error {
	if err := s.run(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Server) run(ctx context.Context) error {
	cat := catalog.New()
	err := cat.LoadFromProvidersConfig(s.config.ProvidersConfig)
	if err != nil {
		return err
	}

	endpointsServer, err := s.newEndpointsServer(cat)
	if err != nil {
		return err
	}

	err = util.RunTasks(ctx, endpointsServer.ListenAndServe)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (s *Server) newEndpointsServer(catalog catalog.Catalog) (endpoints.Server, error) {
	config := &endpoints.Config{
		TCPAddress:          s.config.TCPAddress,
		LocalAddress:        s.config.LocalAddress,
		DatastoreConnString: s.config.DBConnString,
		Logger:              s.config.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
		Catalog:             catalog,
	}

	return endpoints.New(config)
}
