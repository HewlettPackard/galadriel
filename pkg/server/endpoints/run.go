package endpoints

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
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
	err := util.RunTasks(ctx,
		e.runTCPServer,
		e.runUDSServer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (e *Endpoints) runTCPServer(ctx context.Context) error {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	e.addTCPHandlers(server)

	server.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return e.validateToken(c, key)
	}))

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
	server := &http.Server{}

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

	e.addHandlers()

	e.Logger.Infof("Starting UDS Server on %s", e.LocalAddr.String())
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(l)
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

func (e *Endpoints) addHandlers() {
	http.HandleFunc("/createTrustDomain", e.createTrustDomainHandler)
	http.HandleFunc("/listTrustDomains", e.listTrustDomainsHandler)
	http.HandleFunc("/createRelationship", e.createRelationshipHandler)
	http.HandleFunc("/listRelationships", e.listRelationshipsHandler)
	http.HandleFunc("/generateToken", e.generateTokenHandler)
}

func (e *Endpoints) addTCPHandlers(server *echo.Echo) {
	server.CONNECT("/onboard", e.onboardHandler)
	server.POST("/bundle", e.postBundleHandler)
	server.POST("/bundle/sync", e.syncFederatedBundleHandler)
}
