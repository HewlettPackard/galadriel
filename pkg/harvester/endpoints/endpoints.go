package endpoints

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type Endpoints struct {
	localAddress net.Addr
	client       galadrielclient.Client
	logger       logrus.FieldLogger
}

// Config represents the configuration of the Harvester Endpoints.
type Config struct {
	LocalAddress net.Addr // UDS socket address the Harvester will listen on
	Client       galadrielclient.Client
	Logger       logrus.FieldLogger
}

func New(cfg *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(cfg.LocalAddress); err != nil {
		return nil, err
	}

	return &Endpoints{
		localAddress: cfg.LocalAddress,
		client:       cfg.Client,
		logger:       cfg.Logger,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
	e.logger.Debug("Initializing API endpoints")
	err := util.RunTasks(ctx,
		e.startUDSListener,
	)
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	return err
}

func (e *Endpoints) startUDSListener(ctx context.Context) error {
	server := echo.New()

	l, err := net.Listen(e.localAddress.Network(), e.localAddress.String())
	if err != nil {
		return fmt.Errorf("error listening on UDS: %w", err)
	}
	defer l.Close()

	e.addUDSHandlers(server)

	log := e.logger.WithFields(logrus.Fields{
		telemetry.Network: e.localAddress.Network(),
		telemetry.Address: e.localAddress.String()})

	errChan := make(chan error)
	go func() {
		log.Info("Started UDS listener")
		errChan <- server.Server.Serve(l)
	}()

	select {
	case err := <-errChan:
		log.WithError(err).Error("UDS listener stopped prematurely")
		return err
	case <-ctx.Done():
		e.logger.Info("Stopping UDS listener")
		err := server.Close()
		if err != nil {
			log.WithError(err).Error("Error closing UDS listener")
		}

		log.Info("UDS listener stopped")
		return nil
	}
}

func (e *Endpoints) addUDSHandlers(server *echo.Echo) {
	admin.RegisterHandlers(server, NewAdminAPIHandlers(e.logger, e.client))
}
