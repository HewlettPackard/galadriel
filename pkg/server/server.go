package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/HewlettPackard/galadriel/pkg/server/endpoints"
	"github.com/google/uuid"
)

// TODO: consider making this a configuration option
const defaultKeyType = cryptoutil.RSA2048

// Server represents a Galadriel Server.
type Server struct {
	config *Config
}

// New creates a new instance of the Galadriel Server.
func New(config *Config) *Server {
	return &Server{config: config}
}

// Run starts running the Galadriel Server, starting its endpoints.
func (s *Server) Run(ctx context.Context) error {
	// TODO: this method doesn't add any logic, move run() logic here
	if err := s.run(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Server) run(ctx context.Context) error {
	cat := catalog.New()
	err := cat.LoadFromProvidersConfig(s.config.ProvidersConfig)
	if err != nil {
		return fmt.Errorf("failed to load catalogs from providers config: %w", err)
	}

	// TODO: consider moving the datastore to the catalog?
	ds, err := datastore.NewSQLDatastore(s.config.Logger, s.config.DBConnString)
	if err != nil {
		return fmt.Errorf("failed to create datastore: %w", err)
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

	endpointsServer, err := s.newEndpointsServer(cat, ds, jwtIssuer, jwtValidator)
	if err != nil {
		return fmt.Errorf("failed to create endpoints server: %w", err)
	}

	err = util.RunTasks(ctx, endpointsServer.ListenAndServe)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func (s *Server) newEndpointsServer(catalog catalog.Catalog, ds datastore.Datastore, jwtIssuer jwt.Issuer, jwtValidator jwt.Validator) (endpoints.Server, error) {
	config := &endpoints.Config{
		TCPAddress:   s.config.TCPAddress,
		LocalAddress: s.config.LocalAddress,
		Logger:       s.config.Logger.WithField(telemetry.SubsystemName, telemetry.Endpoints),
		Datastore:    ds,
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
	key, err := keyManager.GenerateKey(ctx, keyID.String(), defaultKeyType)
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
