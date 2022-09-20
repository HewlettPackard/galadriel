package endpoints

import (
	"context"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
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

func (e *Endpoints) addHandlers(ctx context.Context) {
	e.generateTokenHandler(ctx)
}

func (e *Endpoints) generateTokenHandler(ctx context.Context) {
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {

		token, err := util.GenerateToken()
		if err != nil {
			e.Log.Errorf("failed to generate token: %v", err)
			w.WriteHeader(500)
			return
		}

		t := &datastore.JoinToken{
			Token:  token,
			Expiry: time.Now(),
		}

		err = e.DataStore.CreateJoinToken(ctx, t)
		if err != nil {
			_, _ = io.WriteString(w, "failed to generate token")
			w.WriteHeader(500)
			return
		}

		_, err = io.WriteString(w, t.Token)
		if err != nil {
			e.Log.Errorf("failed to return token: %v", err)
			w.WriteHeader(500)
			return
		}
	})
}
