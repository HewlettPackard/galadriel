package endpoints

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

const (
	// owner has full access over the path, others can read and list
	udsDirPermissions = 0755
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
	DataStore  datastore.DataStore
	Log        logrus.FieldLogger
}

func New(c *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}
	return &Endpoints{
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		DataStore:  c.Catalog.GetDataStore(),
		Log:        c.Log,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
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

func (e *Endpoints) runTCPServer(ctx context.Context) error {
	server := echo.New()

	e.addTCPHandlers(server)

	server.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return e.validateToken(key)
	}))

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

func (e *Endpoints) runUDSServer(ctx context.Context) error {
	server := &http.Server{}

	udsDir := filepath.Dir(e.LocalAddr.String())
	if err := os.MkdirAll(udsDir, udsDirPermissions); err != nil {
		return fmt.Errorf("failed to create local UDS directory %s: %v", udsDir, err)
	}

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on UDS: %w", err)
	}
	defer l.Close()

	e.addHandlers()

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

func (e *Endpoints) addHandlers() {
	http.HandleFunc("/createMember", e.createMemberHandler)
	http.HandleFunc("/listMembers", e.listMembersHandler)
	http.HandleFunc("/createRelationship", e.createRelationshipHandler)
	http.HandleFunc("/listRelationships", e.listRelationshipsHandler)
	http.HandleFunc("/generateToken", e.generateTokenHandler)
}

func (e *Endpoints) addTCPHandlers(server *echo.Echo) {
	server.CONNECT("/onboard", e.onboardHandler)
}
