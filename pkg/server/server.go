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
	cat, err := catalog.Load(ctx, catalog.Config{Logger: s.config.Logger})
	if err != nil {
		return err
	}
	defer cat.Close()

	endpointsServer, err := s.newEndpointsServer(cat)
	if err != nil {
		return err
	}

	metrics, err := telemetry.NewMetrics(&telemetry.MetricsConfig{
		Logger:      s.config.Logger.WithField(telemetry.SubsystemName, telemetry.Telemetry),
		ServiceName: telemetry.GaladrielServer,
	})
	if err != nil {
		return err
	}

	err = util.RunTasks(ctx, endpointsServer.ListenAndServe, metrics.ListenAndServe)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (s *Server) newEndpointsServer(cat catalog.Catalog) (endpoints.Server, error) {
	config := &endpoints.Config{
		TCPAddress:   s.config.TCPAddress,
		LocalAddress: s.config.LocalAddress,
		Catalog:      cat,
		Logger:       s.config.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
	}

	return endpoints.New(config)
}
