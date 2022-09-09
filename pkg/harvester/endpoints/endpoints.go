package endpoints

import (
	"context"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type Endpoints struct {
	TCPAddress   *net.TCPAddr
	LocalAddress net.Addr
	Log          logrus.FieldLogger
}

func New(c Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}

	return &Endpoints{
		TCPAddress:   c.TCPAddress,
		LocalAddress: c.LocalAddress,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
	l, err := net.Listen(e.LocalAddress.Network(), e.LocalAddress.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

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
