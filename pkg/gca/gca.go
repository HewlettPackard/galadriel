package gca

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/gca/endpoints"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
)

// clock used for time operations that allows to use a Fake for testing
var clk = clock.New()

// Config conveys the configuration for the Galadriel serverCA.
type Config struct {
	// Address of the Galadriel serverCA
	TCPAddress *net.TCPAddr

	// Address of the Galadriel serverCA to be reached locally
	LocalAddress net.Addr

	// Path to the Galadriel serverCA Root Cert File
	RootCertPath string

	// Path to the Galadriel serverCA Private Key File
	RootKeyPath string

	Logger logrus.FieldLogger

	// jwtTokenTTL of the X509 certificates provided by this GCA
	X509CertTTL time.Duration

	// jwtTokenTTL of the JWT tokens provided by this GCA
	JWTCertTTL time.Duration
}

// GCA is a struct that represents a Galadriel serverCA.
type GCA struct {
	config   *Config
	serverCA ca.ServerCA
}

// New creates a new Galadriel serverCA GCA with the given configuration.
func New(config *Config) (*GCA, error) {
	cert, err := cryptoutil.LoadCertificate(config.RootCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed loading serverCA root certificate: %w", err)
	}
	key, err := cryptoutil.LoadRSAPrivateKey(config.RootKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed loading serverCA root private: %w", err)
	}

	CA, err := ca.New(&ca.Config{
		RootCert: cert,
		RootKey:  key,
		Clock:    clk,
		Logger:   config.Logger.WithField(telemetry.SubsystemName, telemetry.ServerCA),
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating GCA serverCA: %w", err)
	}

	return &GCA{
		config:   config,
		serverCA: CA,
	}, nil
}

// Run starts running the Galadriel GCA, starting its endpoints.
func (g *GCA) Run(ctx context.Context) error {
	if err := g.run(ctx); err != nil {
		return err
	}
	return nil
}

func (g *GCA) run(ctx context.Context) error {
	endpointsServer, err := g.newEndpointsServer()
	if err != nil {
		return err
	}

	err = util.RunTasks(ctx, endpointsServer.ListenAndServe)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (g *GCA) newEndpointsServer() (endpoints.Server, error) {
	config := &endpoints.Config{
		TCPAddress:   g.config.TCPAddress,
		LocalAddress: g.config.LocalAddress,
		Logger:       g.config.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
		ServerCA:     g.serverCA,
		JWTTokenTTL:  g.config.JWTCertTTL,
		X509CertTTL:  g.config.X509CertTTL,
		Clock:        clk,
	}

	return endpoints.New(config)
}

func (g *GCA) Stop() {
}
