// TODO: rename this file to endpoints.go
package endpoints

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

const (
	defaultTTL = 1 * time.Hour
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type Endpoints struct {
	// TODO: unexport these fields
	TCPAddress *net.TCPAddr
	LocalAddr  net.Addr
	Datastore  datastore.Datastore
	Logger     logrus.FieldLogger

	x509CA     x509ca.X509CA
	certsStore *certificateSource

	hooks struct {
		// test hook used to signal that TCP listener is ready
		tcpListening chan struct{}
	}
}

type certificateSource struct {
	mu   sync.RWMutex
	cert *tls.Certificate
}

func New(c *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}

	return &Endpoints{
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		Datastore:  c.Datastore,
		Logger:     c.Logger,
		x509CA:     c.Catalog.GetX509CA(),
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

func (e *Endpoints) runTCPServer(ctx context.Context) error {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	e.addTCPHandlers(server)

	server.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return e.validateToken(c, key)
	}))

	cert, err := e.getTLSCertificate(ctx)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	e.certsStore = &certificateSource{cert: cert}

	tlsConfig := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return e.certsStore.getTLSCertificate(), nil
		},
	}

	httpServer := http.Server{
		Addr:      e.TCPAddress.String(),
		Handler:   server, // set Echo as handler
		TLSConfig: tlsConfig,
	}

	e.Logger.Infof("Starting secure Galadriel Server TCP listening on %s", e.TCPAddress.String())
	errChan := make(chan error)
	go func() {
		e.triggerListeningHook()
		// certificate and key are embedded in the TLS config
		errChan <- httpServer.ListenAndServeTLS("", "")
	}()

	go e.startTLSCertificateRotation(ctx, errChan)

	select {
	case err = <-errChan:
		e.Logger.WithError(err).Error("TCP Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Logger.Info("Stopping TCP Server")
		err = httpServer.Close()
		if err != nil {
			e.Logger.WithError(err).Error("Error closing HTTP TCP Server")
		}
		err = server.Close()
		if err != nil {
			e.Logger.WithError(err).Error("Error closing Echo Server")
		}
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

func (t *certificateSource) setTLSCertificate(cert *tls.Certificate) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cert = cert
}

func (t *certificateSource) getTLSCertificate() *tls.Certificate {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cert
}

func (e *Endpoints) startTLSCertificateRotation(ctx context.Context, errChan chan error) {
	e.Logger.Info("Starting TLS certificate rotator")

	// Start a ticker that rotates the certificate every default interval
	certRotationInterval := defaultTTL / 2
	ticker := time.NewTicker(certRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.Logger.Info("Rotating GCA TLS certificate")
			cert, err := e.getTLSCertificate(ctx)
			if err != nil {
				errChan <- fmt.Errorf("failed to rotate GCA TLS certificate: %w", err)
			}
			e.certsStore.setTLSCertificate(cert)
		case <-ctx.Done():
			e.Logger.Info("Stopped GCA TLS certificate rotator")
			return
		}
	}
}

func (e *Endpoints) getTLSCertificate(ctx context.Context) (*tls.Certificate, error) {
	privateKey, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	params := &x509ca.X509CertificateParams{
		Subject: pkix.Name{
			CommonName: constants.GaladrielServerName,
		},
		TTL:       defaultTTL,
		PublicKey: privateKey.Public(),
		DNSNames:  []string{constants.GaladrielServerName},
	}
	cert, err := e.x509CA.IssueX509Certificate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to issue TLS certificate: %w", err)
	}

	certPEM := cryptoutil.EncodeCertificate(cert[0])
	keyPEM := cryptoutil.EncodeRSAPrivateKey(privateKey.(*rsa.PrivateKey))

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
