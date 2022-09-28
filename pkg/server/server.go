package server

import (
	"context"
	"errors"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	admin "github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
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

func (s *Server) run(ctx context.Context) (err error) {
	cat, err := catalog.Load(ctx, catalog.Config{Log: s.config.Log})
	if err != nil {
		return err
	}
	defer cat.Close()
	harvesterServer := harvester.NewServer(harvester.Config{
		TCPAddress: s.config.TCPAddress,
		Logger:     s.config.Log,
	}, cat)

	adminServer := admin.NewServer(admin.Config{
		LocalAddress: s.config.LocalAddress,
		Logger:       s.config.Log,
	}, cat)

	tasks := []func(context.Context) error{
		harvesterServer.Start,
		adminServer.Start,
	}

	err = util.RunTasks(ctx, tasks)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}
