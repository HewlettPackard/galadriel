package endpoints

import (
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

const (
	// TTL of the Galadriel Server certificate
	certTTL = 2 * time.Hour

	// serverName is used as Common Name and DNSName in the Galadriel Server certificate
	serverName   = "galadriel-server"
	organization = "galadriel"
)

var (
	// Rotation interval of the Galadriel Server certificate
	certRotationInterval = certTTL / 2
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type Endpoints struct {
	CA         *ca.CA
	TCPAddress *net.TCPAddr
	LocalAddr  net.Addr
	Datastore  datastore.Datastore
	Logger     logrus.FieldLogger

	certsStore *tlsCertSource

	hooks struct {
		// test hook used to indicate that is listening on TCP
		tcpListening chan struct{}
	}
}

type tlsCertSource struct {
	mu   sync.RWMutex
	cert *tls.Certificate
}

func New(c *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}

	ds, err := datastore.NewSQLDatastore(c.Logger, c.DatastoreConnString)
	if err != nil {
		return nil, err
	}

	return &Endpoints{
		CA:         c.CA,
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		Datastore:  ds,
		Logger:     c.Logger,
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
	s := echo.New()
	s.HideBanner = true
	s.HidePort = true

	e.addTCPHandlers(s)

	s.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return e.validateToken(c, key)
	}))

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

	s.TLSServer.TLSConfig = tlsConfig
	l, err := net.Listen("tcp", e.TCPAddress.String())
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(l, tlsConfig)

	e.Logger.Infof("Starting secure TCP Server on %s", e.TCPAddress.String())
	errChan := make(chan error)
	go func() {
		e.triggerListeningHook()
		// certificate and key are embedded in the listener TLS config
		errChan <- s.Server.Serve(tlsListener)
	}()

	go e.tlsCertificateRotator(ctx, errChan)

	select {
	case err = <-errChan:
		e.Logger.WithError(err).Error("TCP Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Logger.Info("Stopping TCP Server")
		s.Close()
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
	server.POST("/onboard", e.onboardHandler)
	server.POST("/bundle", e.postBundleHandler)
	server.POST("/bundle/sync", e.syncFederatedBundleHandler)
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

func (e *Endpoints) tlsCertificateRotator(ctx context.Context, errChan chan error) {
	e.Logger.Info("Starting GS TLS certificate rotator")

	// Start a ticker that rotates the certificate every default interval
	ticker := time.NewTicker(certRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.Logger.Info("Rotating GS TLS certificate")
			cert, err := e.getTLSCertificate(ctx)
			if err != nil {
				errChan <- fmt.Errorf("failed to rotate GCA TLS certificate: %w", err)
			}
			e.certsStore.setCert(cert)
		case <-ctx.Done():
			e.Logger.Info("Stopped GS TLS certificate rotator")
			return
		}
	}
}

func (e *Endpoints) triggerListeningHook() {
	if e.hooks.tcpListening != nil {
		e.hooks.tcpListening <- struct{}{}
	}
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
