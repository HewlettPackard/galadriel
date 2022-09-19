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

func New(c Config) (*Endpoints, error) {
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
	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

	e.addHandlers(ctx)

	localServer := &http.Server{}
	tcpServer := echo.New()

	errLocalServer := make(chan error)
	go func() {
		errLocalServer <- localServer.Serve(l)
	}()

	errTcpServer := make(chan error)
	go func() {
		errTcpServer <- tcpServer.Start(e.TCPAddress.String())
	}()

	select {
	case err = <-errLocalServer:
	case <-ctx.Done():
		if err != nil {
			fmt.Printf("error serving HTTP on uds: %v", err)
		}
		e.Log.Println("Stopping HTTP Server")
		localServer.Close()
		tcpServer.Close()
	}

	return nil
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
