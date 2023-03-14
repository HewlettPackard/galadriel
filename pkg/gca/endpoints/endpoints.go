package endpoints

import (
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/gca/endpoints/jwt"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
)

const (
	// TTL of the Galadriel CA certificate
	certTTL = 2 * time.Hour

	// serverName is used as Common Name and DNSName in the Galadriel CA certificate
	serverName   = "galadriel-ca"
	organization = "galadriel"
)

var (
	// Rotation interval of the Galadriel CA certificate
	certRotationInterval = certTTL / 2
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

// Config represents the configuration of the Galadriel CA Endpoints
type Config struct {
	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// LocalAddress is the local address to bind the listener to.
	LocalAddress net.Addr

	Logger logrus.FieldLogger

	// CA is used for signing X.509 certificates and JWTs
	CA *ca.CA

	JWTTokenTTL time.Duration
	X509CertTTL time.Duration

	Clock clock.Clock
}

type Endpoints struct {
	CA         *ca.CA
	TCPAddress *net.TCPAddr
	LocalAddr  net.Addr
	Logger     logrus.FieldLogger
	Clock      clock.Clock
	config     *Config

	certsStore *tlsCertSource

	hooks struct {
		// test hook used to indicate that is listening on TCP
		tcpListening chan struct{}
	}
}

func New(c *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}

	return &Endpoints{
		CA:         c.CA,
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		Logger:     c.Logger,
		Clock:      c.Clock,
		config:     c,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
	err := util.RunTasks(ctx,
		e.runTCPServer,
		e.runUDSServer,
	)
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	return err
}

type tlsCertSource struct {
	mu   sync.RWMutex
	cert *tls.Certificate
}

func (t *tlsCertSource) setCert(cert *tls.Certificate) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cert = cert
}

func (t *tlsCertSource) getCert() *tls.Certificate {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cert
}

func (e *Endpoints) runTCPServer(ctx context.Context) error {
	var err error

	cert, err := e.getTLSCertificate(ctx)
	if err != nil {
		return err
	}

	e.certsStore = &tlsCertSource{cert: cert}

	tlsConfig := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return e.certsStore.getCert(), nil
		},
	}

	// TCP API handlers
	mux := http.NewServeMux()

	jwtHandler, err := jwt.NewHandler(&jwt.Config{
		CA:          e.CA,
		Logger:      e.Logger,
		JWTTokenTTL: e.config.JWTTokenTTL,
		Clock:       e.Clock,
	})
	if err != nil {
		return fmt.Errorf("failed to build JWT handler")
	}
	mux.Handle("/jwt", jwtHandler)

	server := &http.Server{
		Addr:      e.TCPAddress.String(),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	e.Logger.Infof("Starting TCP GCA on %s", e.TCPAddress.String())
	errChan := make(chan error)
	go func() {
		e.triggerListeningHook()
		// certificate and key are embedded in the TLS config
		errChan <- server.ListenAndServeTLS("", "")
	}()

	go e.tlsCertificateRotator(ctx, errChan)

	select {
	case err = <-errChan:
		e.Logger.WithError(err).Error("TCP GCA stopped prematurely")
		return err
	case <-ctx.Done():
		e.Logger.Info("Stopping TCP GCA")
		server.Close()
		<-errChan
		e.Logger.Info("TCP GCA stopped")
		return nil
	}
}

func (e *Endpoints) runUDSServer(ctx context.Context) error {
	os.Remove(e.LocalAddr.String())

	// UDS API handlers
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error tcpListening on uds: %w", err)
	}
	defer l.Close()

	e.Logger.Infof("Starting UDS GCA on %s", e.LocalAddr.String())
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(l)
	}()

	select {
	case err = <-errChan:
		e.Logger.WithError(err).Error("Local GCA stopped prematurely")
		return err
	case <-ctx.Done():
		e.Logger.Info("Stopping UDS GCA")
		server.Close()
		<-errChan
		e.Logger.Info("UDS GCA stopped")

		return nil
	}
}

func (e *Endpoints) tlsCertificateRotator(ctx context.Context, errChan chan error) {
	e.Logger.Info("Starting GCA certificate rotator")

	// Start a ticker that rotates the certificate every default interval
	ticker := time.NewTicker(certRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.Logger.Info("Rotating GCA CACertificate")
			cert, err := e.getTLSCertificate(ctx)
			if err != nil {
				errChan <- fmt.Errorf("failed to rotate GCA certificate: %w", err)
			}
			e.certsStore.setCert(cert)
		case <-ctx.Done():
			e.Logger.Info("Stopped GCA CACertificate rotator")
			return
		}
	}
}

func (e *Endpoints) getTLSCertificate(ctx context.Context) (*tls.Certificate, error) {
	privateKey, err := cryptoutil.CreateRSAKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	params := ca.X509CertificateParams{
		PublicKey: privateKey.Public(),
		TTL:       certTTL,
		Subject: pkix.Name{
			CommonName:   serverName,
			Organization: []string{organization},
		},
	}
	cert, err := e.CA.SignX509Certificate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := cryptoutil.EncodeCertificate(cert)
	keyPEM := cryptoutil.EncodeRSAPrivateKey(privateKey)

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &certificate, nil
}

func (e *Endpoints) triggerListeningHook() {
	if e.hooks.tcpListening != nil {
		e.hooks.tcpListening <- struct{}{}
	}
}
