package endpoints

import (
	"context"
	"fmt"
	"net"
	"net/http"
	
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type EndpointHandler struct {
	TCPAddress *net.TCPAddr
	LocalAddr  net.Addr
	DataStore  datastore.DataStore
	Log        logrus.FieldLogger
}

func New(c *Config) (*EndpointHandler, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}
	return &EndpointHandler{
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		DataStore:  c.Catalog.GetDataStore(),
		Log:        c.Log,
	}, nil
}

func (e *EndpointHandler) ListenAndServe(ctx context.Context) error {
	tasks := []func(context.Context) error{
		e.runTCPServer,
		e.runUDSServer,
	}

	err := util.RunTasks(ctx, tasks)
	if err != nil {
		return err
	}

	return nil
}

func (e *EndpointHandler) runTCPServer(ctx context.Context) error {
	server := echo.New()

	e.Log.Info("Starting TCP Server")
	errChan := make(chan error)
	go func() {
		errChan <- server.Start(e.TCPAddress.String())
	}()

	var err error
	select {
	case err = <-errChan:
		e.Log.WithError(err).Error("TCP Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Log.Info("Stopping TCP Server")
		server.Close()
		<-errChan
		e.Log.Info("TCP Server stopped")
		return nil
	}
}

func (e *EndpointHandler) runUDSServer(ctx context.Context) error {
	server := &http.Server{}

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

	e.addHandlers(ctx)

	e.Log.Info("Starting UDS Server")
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(l)
	}()

	select {
	case err = <-errChan:
		e.Log.WithError(err).Error("Local Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Log.Info("Stopping UDS Server")
		server.Close()
		<-errChan
		e.Log.Info("UDS Server stopped")
		return nil
	}
}

func (e *EndpointHandler) addHandlers(ctx context.Context) {
	e.createMemberHandler(ctx)
	e.createRelationshipHandler(ctx)
	e.generateTokenHandler(ctx)
}
