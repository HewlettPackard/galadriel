package endpoints

import (
	"context"
	"fmt"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"

	adminAPI "github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	harvesterapi "github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type Endpoints struct {
	TCPAddress *net.TCPAddr
	LocalAddr  net.Addr
	Datastore  datastore.Datastore
	Logger     logrus.FieldLogger
}

func New(c *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}

	ds, err := datastore.NewSQLDatastore(c.Logger, c.DatastoreConnString)
	if err != nil {
		return nil, err
	}

	return &Endpoints{
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		Datastore:  ds,
		Logger:     c.Logger,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
	if err := util.RunTasks(ctx, e.runTCPServer, e.runUDSServer); err != nil {
		return err
	}
	return nil
}

func (e *Endpoints) runTCPServer(ctx context.Context) error {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	e.addTCPHandlers(server)
	e.addTCPMiddlewares(server)

	e.Logger.Infof("Starting TCP Server on %s", e.TCPAddress.String())
	errChan := make(chan error)
	go func() {
		errChan <- server.Start(e.TCPAddress.String())
	}()

	var err error
	select {
	case err = <-errChan:
		e.Logger.WithError(err).Error("TCP Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Logger.Info("Stopping TCP Server")
		server.Close()
		<-errChan
		e.Logger.Info("TCP Server stopped")
		return nil
	}
}

func (e *Endpoints) runUDSServer(ctx context.Context) error {
	server := echo.New()

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

	e.addUDSHandlers(server)

	e.Logger.Infof("Starting UDS Server on %s", e.LocalAddr.String())
	errChan := make(chan error)
	go func() {
		errChan <- server.Server.Serve(l)
	}()

	select {
	case err = <-errChan:
		e.Logger.WithError(err).Error("Local Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Logger.Info("Stopping UDS Server")
		server.Close()
		<-errChan
		e.Logger.Info("UDS Server stopped")
		return nil
	}
}

func (e *Endpoints) addUDSHandlers(server *echo.Echo) {
	logger := e.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints)
	adminAPI.RegisterHandlers(server, NewAdminAPIHandlers(logger, e.Datastore))
}

func (e *Endpoints) addTCPHandlers(server *echo.Echo) {
	logger := e.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints)
	harvesterAPI.RegisterHandlers(server, NewHarvesterAPIHandlers(logger, e.Datastore))
}

func (e *Endpoints) addTCPMiddlewares(server *echo.Echo) {
	logger := e.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints)
	authNMiddleware := NewAuthenticationMiddleware(logger, e.Datastore)
	server.Use(middleware.KeyAuth(authNMiddleware.Authenticate))
}
