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
	"strings"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	adminapi "github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	harvesterapi "github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

const (
	serverCertificateTTL = 1 * time.Hour
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
	Datastore  db.Datastore
	Logger     logrus.FieldLogger

	x509CA       x509ca.X509CA
	jwtIssuer    jwt.Issuer
	jwtValidator jwt.Validator
	certsStore   *certificateSource

	hooks struct {
		// test hook used to signal that TCP listener is ready
		tcpListening chan struct{}
	}
}

// Config represents the configuration of the Galadriel Server Endpoints
type Config struct {
	TCPAddress   *net.TCPAddr
	LocalAddress net.Addr
	JWTIssuer    jwt.Issuer
	JWTValidator jwt.Validator
	Catalog      catalog.Catalog
	Logger       logrus.FieldLogger
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
		TCPAddress:   c.TCPAddress,
		LocalAddr:    c.LocalAddress,
		Datastore:    c.Catalog.GetDatastore(),
		Logger:       c.Logger,
		x509CA:       c.Catalog.GetX509CA(),
		jwtIssuer:    c.JWTIssuer,
		jwtValidator: c.JWTValidator,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
	e.Logger.Debug("Initializing API endpoints")
	err := util.RunTasks(ctx,
		e.startTCPListener,
		e.startUDSListener,
	)
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	return err
}

func (e *Endpoints) startTCPListener(ctx context.Context) error {
	e.Logger.Debug("Starting TCP listener")

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	e.addTCPHandlers(server)
	e.addTCPMiddlewares(server)

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

	log := e.Logger.WithFields(logrus.Fields{
		telemetry.Network: e.TCPAddress.Network(),
		telemetry.Address: e.TCPAddress.String()})

	errChan := make(chan error)
	go func() {
		e.triggerListeningHook()
		log.Info("Started TCP listener")
		// certificate and key are embedded in the TLS config
		errChan <- httpServer.ListenAndServeTLS("", "")
	}()

	go e.startTLSCertificateRotation(ctx, errChan)

	select {
	case err := <-errChan:
		log.WithError(err).Error("TCP listener stopped prematurely")
		return err
	case <-ctx.Done():
		log.Info("Stopping TCP listener")
		err = httpServer.Close()
		if err != nil {
			log.WithError(err).Error("Error closing TCP listener")
		}
		err = server.Close()
		if err != nil {
			e.Logger.WithError(err).Error("Error closing Echo Server")
		}
		<-errChan
		log.Info("TCP listener stopped")
		return nil
	}
}

func (e *Endpoints) startUDSListener(ctx context.Context) error {
	e.Logger.Debug("Starting UDS listener")
	server := echo.New()

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on UDS: %w", err)
	}
	defer l.Close()

	e.addUDSHandlers(server)

	log := e.Logger.WithFields(logrus.Fields{
		telemetry.Network: e.LocalAddr.Network(),
		telemetry.Address: e.LocalAddr.String()})

	errChan := make(chan error)
	go func() {
		log.Info("Started UDS listener")
		errChan <- server.Server.Serve(l)
	}()

	select {
	case err = <-errChan:
		log.WithError(err).Error("Local listener stopped prematurely")
		return err
	case <-ctx.Done():
		log.Info("Stopping UDS listener")
		err := server.Close()
		if err != nil {
			log.WithError(err).Error("Error closing UDS listener")
		}
		<-errChan
		log.Info("UDS listener stopped")
		return nil
	}
}

func (e *Endpoints) addUDSHandlers(server *echo.Echo) {
	adminapi.RegisterHandlers(server, NewAdminAPIHandlers(e.Logger, e.Datastore))
}

func (e *Endpoints) addTCPHandlers(server *echo.Echo) {
	harvesterapi.RegisterHandlers(server, NewHarvesterAPIHandlers(e.Logger, e.Datastore, e.jwtIssuer, e.jwtValidator))
}

func (e *Endpoints) addTCPMiddlewares(server *echo.Echo) {
	logger := e.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints)
	authNMiddleware := NewAuthenticationMiddleware(logger, e.Datastore, e.jwtValidator)

	skipOnboard := func(c echo.Context) bool {
		return strings.Contains(c.Request().URL.Path, "/onboard")
	}

	myMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipOnboard(c) {
				return next(c)
			}
			return middleware.KeyAuth(authNMiddleware.Authenticate)(next)(c)
		}
	}

	server.Use(myMiddleware, middleware.Recover(), middleware.CORS())
}

func (t *certificateSource) setTLSCertificate(cert *tls.Certificate) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cert = cert
}

func (t *certificateSource) getTLSCertificate() *tls.Certificate {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cert
}

func (e *Endpoints) startTLSCertificateRotation(ctx context.Context, errChan chan error) {
	e.Logger.Info("Started TLS certificate rotator")

	// Start a ticker that rotates the certificate every default interval
	certRotationInterval := serverCertificateTTL / 2
	ticker := time.NewTicker(certRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.Logger.Debug("Rotating Server TLS certificate")
			cert, err := e.getTLSCertificate(ctx)
			if err != nil {
				errChan <- fmt.Errorf("failed to rotate Server TLS certificate: %w", err)
			}
			e.certsStore.setTLSCertificate(cert)
		case <-ctx.Done():
			e.Logger.Info("Stopped Server TLS certificate rotator")
			return
		}
	}
}

func (e *Endpoints) getTLSCertificate(ctx context.Context) (*tls.Certificate, error) {
	privateKey, err := cryptoutil.GenerateSigner(cryptoutil.DefaultKeyType)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	params := &x509ca.X509CertificateParams{
		Subject: pkix.Name{
			CommonName: constants.GaladrielServerName,
		},
		TTL:       serverCertificateTTL,
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
