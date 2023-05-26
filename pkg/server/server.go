package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/HewlettPackard/galadriel/pkg/server/endpoints"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Server represents a Galadriel Server.
type Server struct {
	config *Config
}

// Config conveys configurations for the Galadriel Server
type Config struct {
	TCPAddress      *net.TCPAddr
	LocalAddress    net.Addr
	Logger          logrus.FieldLogger
	ProvidersConfig *catalog.ProvidersConfig
}

// New creates a new instance of the Galadriel Server.
func New(config *Config) *Server {
	return &Server{config: config}
}

// Run starts the Galadriel Server, initializing the components and listening for incoming requests.
// It performs the following steps:
// 1. Loads catalogs from the providers configuration.
// 2. Creates a JWT issuer based on the key manager from the catalogs.
// 3. Sets up a JWT validator.
// 4. Creates the endpoints server, which handles incoming requests.
// 5. Starts the endpoints server and listens for requests until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	s.config.Logger.Info("Starting Galadriel Server")

	cat := catalog.New()
	err := cat.LoadFromProvidersConfig(s.config.ProvidersConfig)
	if err != nil {
		return fmt.Errorf("failed to load catalogs from providers config: %w", err)
	}

	jwtIssuer, err := s.createJWTIssuer(ctx, cat.GetKeyManager())
	if err != nil {
		return fmt.Errorf("failed to create JWT issuer: %w", err)
	}

	c := &jwt.ValidatorConfig{
		KeyManager:       cat.GetKeyManager(),
		ExpectedAudience: []string{constants.GaladrielServerName},
	}
	jwtValidator := jwt.NewDefaultJWTValidator(c)

	endpointsServer, err := s.newEndpointsServer(cat, jwtIssuer, jwtValidator)
	if err != nil {
		return fmt.Errorf("failed to create endpoints server: %w", err)
	}

	err = util.RunTasks(ctx, endpointsServer.ListenAndServe)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (s *Server) newEndpointsServer(catalog catalog.Catalog, jwtIssuer jwt.Issuer, jwtValidator jwt.Validator) (endpoints.Server, error) {
	config := &endpoints.Config{
		TCPAddress:   s.config.TCPAddress,
		LocalAddress: s.config.LocalAddress,
		Logger:       s.config.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
		Catalog:      catalog,
		JWTIssuer:    jwtIssuer,
		JWTValidator: jwtValidator,
	}

	return endpoints.New(config)
}

func (s *Server) createJWTIssuer(ctx context.Context, keyManager keymanager.KeyManager) (jwt.Issuer, error) {
	keyID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key ID: %v", err)
	}
	key, err := keyManager.GenerateKey(ctx, keyID.String(), cryptoutil.DefaultKeyType)
	if err != nil {
		return nil, err
	}

	jwtIssuer, err := jwt.NewJWTCA(&jwt.Config{
		Signer: key.Signer(),
		Kid:    key.ID(),
	})
	if err != nil {
		return nil, err
	}

	return jwtIssuer, nil
}
