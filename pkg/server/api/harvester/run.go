package harvester

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Server interface {
	Start(context.Context) error
}

type echoServer struct {
	server    *echo.Echo
	DataStore datastore.DataStore
	config    Config
}

func NewServer(config Config, cat *catalog.Repository) Server {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	return &echoServer{
		server:    e,
		config:    config,
		DataStore: cat.GetDataStore(),
	}
}

func (s *echoServer) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- s.server.Start(s.config.TCPAddress.String())
	}()

	s.addHandlers()
	var err error
	select {
	case err = <-errChan:
		s.config.Logger.WithError(err).Error("TCP Server stopped prematurely")
		return err
	case <-ctx.Done():
		s.config.Logger.Info("Stopping TCP Server")
		s.server.Close()
		<-errChan
		s.config.Logger.Info("TCP Server stopped")
		return nil
	}
}

func (s *echoServer) addHandlers() {
	s.server.CONNECT(connectPath, s.onboardHandler)
	s.server.POST(postBundlePath, s.postBundleHandler)
	s.server.POST(syncBundlesPath, s.syncBundleHandler)
}
