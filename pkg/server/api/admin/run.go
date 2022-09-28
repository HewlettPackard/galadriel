package admin

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/sirupsen/logrus"
)

type Server interface {
	Start(context.Context) error
}

type localServer struct {
	config    Config
	DataStore datastore.DataStore
	Logger    logrus.Logger
}

func NewServer(config Config, cat *catalog.Repository) Server {
	return &localServer{
		config:    config,
		DataStore: cat.GetDataStore(),
	}
}

func (s *localServer) Start(ctx context.Context) error {
	s.config.Logger.Info("Starting TCP Server")

	l, err := net.Listen(s.config.LocalAddress.Network(), s.config.LocalAddress.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

	s.addHandlers()

	s.config.Logger.Info("Starting UDS Server")
	errChan := make(chan error)
	httpServer := &http.Server{}
	go func() {
		errChan <- httpServer.Serve(l)
	}()

	select {
	case err = <-errChan:
		s.config.Logger.WithError(err).Error("Local Server stopped prematurely")
		return err
	case <-ctx.Done():
		s.config.Logger.Info("Stopping UDS Server")
		httpServer.Close()
		<-errChan
		s.config.Logger.Info("UDS Server stopped")
		return nil
	}
}

func (s *localServer) addHandlers() {
	http.HandleFunc(CreateMemberPath, s.createMemberHandler)
	http.HandleFunc(CreateRelationshipPath, s.createRelationshipHandler)
	http.HandleFunc(GenerateTokenPath, s.generateTokenHandler)
}
