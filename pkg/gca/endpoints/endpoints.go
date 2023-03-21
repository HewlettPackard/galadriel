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
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
)

const (
	// TTL of the Galadriel serverCA certificate
	certTTL = 2 * time.Hour

	// serverName is used as Common Name and DNSName in the Galadriel serverCA certificate
	serverName   = "galadriel-ca"
	organization = "galadriel"
)

var (
	// Rotation interval of the Galadriel serverCA certificate
	certRotationInterval = certTTL / 2
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

// Config represents the configuration of the Galadriel ServerCA Endpoints
type Config struct {
	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// LocalAddress is the local address to bind the listener to.
	LocalAddress net.Addr

	Logger logrus.FieldLogger

	// ServerCA is used for signing X.509 certificates and JWTs
	ServerCA ca.ServerCA

	JWTTokenTTL time.Duration
	X509CertTTL time.Duration

	Clock clock.Clock
}

type Endpoints struct {
	serverCA   ca.ServerCA
	tcpAddress *net.TCPAddr
	localAddr  net.Addr
	logger     logrus.FieldLogger
	clock      clock.Clock
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
		serverCA:   c.ServerCA,
		tcpAddress: c.TCPAddress,
		localAddr:  c.LocalAddress,
		logger:     c.Logger,
		clock:      c.Clock,
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

	server := &http.Server{
		Addr:      e.tcpAddress.String(),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	e.logger.Infof("Starting secure GCA TCP endpoint listening on %s", e.tcpAddress.String())
	errChan := make(chan error)
	go func() {
		e.triggerListeningHook()
		// certificate and key are embedded in the TLS config
		errChan <- server.ListenAndServeTLS("", "")
	}()

	go e.startTLSCertificateRotation(ctx, errChan)

	select {
	case err = <-errChan:
		e.logger.WithError(err).Error("GCA TCP endpoint stopped prematurely")
		return err
	case <-ctx.Done():
		e.logger.Info("Stopping GCA TCP endpoint")
		err = server.Close()
		if err != nil {
			e.logger.WithError(err).Error("Error closing GCA TCP endpoint")
		}
		<-errChan
		e.logger.Info("GCA TCP endpoint stopped")
		return nil
	}
}

func (e *Endpoints) runUDSServer(ctx context.Context) error {
	os.Remove(e.localAddr.String())

	// UDS API handlers
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}

	l, err := net.Listen(e.localAddr.Network(), e.localAddr.String())
	if err != nil {
		return fmt.Errorf("error tcpListening on uds: %w", err)
	}
	defer l.Close()

	e.logger.Infof("Starting GCA UDS endpoint listening on %s", e.localAddr.String())
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(l)
	}()

	select {
	case err = <-errChan:
		e.logger.WithError(err).Error("Local GCA UDS endpoint stopped prematurely")
		return err
	case <-ctx.Done():
		e.logger.Info("Stopping GCA UDS endpoint")
		err = server.Close()
		if err != nil {
			e.logger.WithError(err).Error("Error closing GCA UDS endpoint")
		}
		<-errChan
		e.logger.Info("GCA UDS endpoint stopped")

		return nil
	}
}

func (e *Endpoints) startTLSCertificateRotation(ctx context.Context, errChan chan error) {
	e.logger.Info("Starting GCA TLS certificate rotator")

	// Start a ticker that rotates the certificate every default interval
	ticker := time.NewTicker(certRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.logger.Info("Rotating GCA TLS certificate")
			cert, err := e.getTLSCertificate(ctx)
			if err != nil {
				errChan <- fmt.Errorf("failed to rotate GCA TLS certificate: %w", err)
			}
			e.certsStore.setCert(cert)
		case <-ctx.Done():
			e.logger.Info("Stopped GCA TLS certificate rotator")
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
	cert, err := e.serverCA.SignX509Certificate(params)
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
