package endpoints

import (
	"context"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/sirupsen/logrus"
	"net"
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
	e.Log.Fatal("not implemented")
	return nil
}
