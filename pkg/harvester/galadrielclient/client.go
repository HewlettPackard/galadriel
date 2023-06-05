package galadrielclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/diskutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	jwtRotationInterval = 5 * time.Minute
	onboardPath         = "/trust-domain/onboard"
	tokenFile           = "jwt-token"
)

var (
	NotOnboardedErr = errors.New("client has not been onboarded to Galadriel Server")
)

// Client represents a client to interact with the Galadriel Server API.
type Client interface {
	SyncBundles(context.Context, []*entity.Bundle) ([]*entity.Bundle, map[spiffeid.TrustDomain][]byte, error)
	PostBundle(context.Context, *entity.Bundle) error
	GetRelationships(context.Context, entity.ConsentStatus) ([]*entity.Relationship, error)
	UpdateRelationship(context.Context, uuid.UUID, entity.ConsentStatus) (*entity.Relationship, error)
}

// Config is a struct that holds the configuration for the Galadriel Server client.
type Config struct {
	TrustDomain            spiffeid.TrustDomain
	GaladrielServerAddress *net.TCPAddr
	TrustBundlePath        string
	DataDir                string
	JoinToken              string
	Logger                 logrus.FieldLogger
}

// client is a struct that implements the Client interface
type client struct {
	client      harvester.ClientInterface
	trustDomain spiffeid.TrustDomain
	jwtStore    *jwtStore
	logger      logrus.FieldLogger
}

// jwtStore is a struct that holds the JWT access token
type jwtStore struct {
	mu            sync.RWMutex
	jwt           string
	tokenFilePath string // File path for storing the JWT token
	logger        logrus.FieldLogger
}

// NewClient creates a new Galadriel Server client, using the given trustBundlePath to validate the server certificate.
// It Onboards the client to the Galadriel Server using the given joinToken.
// If the client has already been onboarded, it will use the existing JWT token.
func NewClient(ctx context.Context, cfg *Config) (Client, error) {
	if cfg.GaladrielServerAddress == nil {
		return nil, errors.New("server address cannot be nil")
	}
	if cfg.TrustBundlePath == "" {
		return nil, errors.New("trust bundle path cannot be empty")
	}
	if _, err := os.Stat(cfg.TrustBundlePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("trust bundle path does not exist: %s", cfg.TrustBundlePath)
	}
	if cfg.DataDir == "" {
		return nil, errors.New("data dir cannot be empty")
	}

	jwtProvider, err := newJwtStore(cfg.DataDir, tokenFile, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT provider: %w", err)
	}

	c, err := createTLSClient(cfg.TrustBundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client for server %s: %w", cfg.GaladrielServerAddress, err)
	}

	serverAddress := fmt.Sprintf("%s://%s", constants.HTTPSScheme, cfg.GaladrielServerAddress.String())

	// Create harvester client
	harvesterClient, err := harvester.NewClient(serverAddress,
		harvester.WithHTTPClient(c),
		harvester.WithRequestEditorFn(createJWTTokenReqEditor(jwtProvider)))
	if err != nil {
		return nil, fmt.Errorf("failed to create harvester client: %w", err)
	}

	client := &client{
		trustDomain: cfg.TrustDomain,
		client:      harvesterClient,
		logger:      cfg.Logger,
		jwtStore:    jwtProvider,
	}

	// if the user provided a join token, try to onboard the Harvester to Galadriel Server
	if cfg.JoinToken != "" {
		if err := client.onboard(ctx, cfg.JoinToken); err != nil {
			return nil, fmt.Errorf("failed to onboard client: %w", err)
		}
	}

	if !client.isClientOnboarded() {
		// this happens if the user did not provide a join token and the Harvester cannot find a stored jwt token
		return nil, errors.New("harvester is not onboarded to Galadriel Server. A join token is required")
	}

	client.logger.Debug("Requesting a new JWT token from Galadriel Server")
	if err := client.getNewJWTToken(ctx); err != nil {
		return nil, fmt.Errorf("could not connect using existing JWT token: %v", err)
	}
	go client.startJWTTokenRotation(ctx)

	return client, nil
}

// GetRelationships retrieves a list of relationships based on the specified consent status.
// It takes the consentStatus parameter, which indicates the desired consent status to filter the relationships.
// If consentStatus is empty, it returns all relationships regardless of consent status.
// The method returns a slice of entity.Relationship representing the filtered relationships.
// If the client is not onboarded, it returns NotOnboardedErr.
// Any other errors encountered during the operation are returned as well.
func (c *client) GetRelationships(ctx context.Context, consentStatus entity.ConsentStatus) ([]*entity.Relationship, error) {
	if c.jwtStore == nil {
		return nil, NotOnboardedErr
	}

	var status api.ConsentStatus
	if consentStatus != "" {
		status = api.ConsentStatus(consentStatus)
	}

	resp, err := c.client.GetRelationships(ctx, c.trustDomain.String(), &harvester.GetRelationshipsParams{ConsentStatus: &status})
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get relationships: %s", string(body))
	}

	var relationships []api.Relationship
	if err := json.Unmarshal(body, &relationships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	// convert relationships to []*entity.Relationship
	var rels []*entity.Relationship
	for _, r := range relationships {
		ent, err := r.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert relationship to entity: %v", err)
		}
		rels = append(rels, ent)
	}

	return rels, nil
}

// UpdateRelationship updates the consent status of a relationship identified by the given relationshipID.
// The client must be onboarded (c.jwtStore != nil) for this operation to succeed.
// It takes the relationshipID parameter, which specifies the ID of the relationship to update,
// and the consentStatus parameter, which indicates the new consent status for the relationship.
// The consentStatus must not be empty, and the relationshipID must not be empty.
// If any of these conditions are not met, the method returns an error.
// If the operation succeeds, it returns nil.
// If the client is not onboarded, it returns NotOnboardedErr.
// Any other errors encountered during the operation are returned as well.
func (c *client) UpdateRelationship(ctx context.Context, relationshipID uuid.UUID, consentStatus entity.ConsentStatus) (*entity.Relationship, error) {
	if c.jwtStore == nil {
		return nil, NotOnboardedErr
	}
	if consentStatus == "" {
		return nil, errors.New("consent status cannot be empty")
	}
	if relationshipID == uuid.Nil {
		return nil, errors.New("relationship id cannot be empty")
	}

	request := harvester.PatchRelationshipRequest{
		ConsentStatus: api.ConsentStatus(consentStatus),
	}

	resp, err := c.client.PatchRelationship(ctx, c.trustDomain.String(), relationshipID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update relationship: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update relationship: %s", string(body))
	}

	var r api.Relationship
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	ent, err := r.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert relationship to entity: %v", err)
	}

	return ent, nil
}

// SyncBundles synchronizes the given bundles with the Galadriel Server. It returns the updated bundles and the
// map of all federated trust domains with active relationships and their bundle digests.
func (c *client) SyncBundles(ctx context.Context, bundles []*entity.Bundle) ([]*entity.Bundle, map[spiffeid.TrustDomain][]byte, error) {
	if c.jwtStore == nil {
		return nil, nil, NotOnboardedErr
	}

	// Create the request body
	digests := make(map[string]string)
	for _, b := range bundles {
		digests[b.TrustDomainName.String()] = util.EncodeToString(b.Digest)
	}
	syncRequest := harvester.PostBundleSyncRequest{
		State: digests,
	}

	resp, err := c.client.BundleSync(ctx, c.trustDomain.String(), syncRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to sync bundles: %s", string(body))
	}

	syncResult := &harvester.PostBundleSyncResponse{}
	if err := json.Unmarshal(body, syncResult); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	var updates []*entity.Bundle
	for td, b := range syncResult.Updates {
		bundle, err := createEntityBundle(td, &b)
		if err != nil {
			return nil, nil, err
		}

		updates = append(updates, bundle)
	}

	state := make(map[spiffeid.TrustDomain][]byte)
	for td, digest := range syncResult.State {
		trustDomain, err := spiffeid.TrustDomainFromString(td)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse trust domain: %w", err)
		}
		d, err := util.DecodeString(digest)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode digest: %w", err)
		}
		state[trustDomain] = d
	}

	return updates, state, nil
}

func (c *client) PostBundle(ctx context.Context, bundle *entity.Bundle) error {
	if c.jwtStore == nil {
		return NotOnboardedErr
	}

	sig := util.EncodeToString(bundle.Signature)
	cert := util.EncodeToString(bundle.SigningCertificate)
	bundlePut := harvester.PutBundleRequest{
		TrustBundle:        string(bundle.Data),
		Digest:             util.EncodeToString(bundle.Digest),
		Signature:          &sig,
		SigningCertificate: &cert,
		TrustDomain:        bundle.TrustDomainName.String(),
	}

	resp, err := c.client.BundlePut(ctx, bundle.TrustDomainName.String(), bundlePut)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post bundle: %s", string(body))
	}

	return nil
}

// isClientOnboarded Check if the client has been onboarded by checking if there is a JWT token
func (c *client) isClientOnboarded() bool {
	return c.jwtStore.getToken() != ""
}

// onboard initiates the onboarding process of the client with the server using the provided token.
// It makes a request to the server with the token and gets a response with a JWT token.
// If the JWT token in the onboard response is empty, an error is returned.
// If the JWT token is valid, it caches in the client jwtStore.
// Finally, it starts the JWT token rotator.
func (c *client) onboard(ctx context.Context, token string) error {
	c.logger.Info("Onboarding Harvester")

	params := harvester.OnboardParams{JoinToken: token}
	resp, err := c.client.Onboard(ctx, c.trustDomain.String(), &params)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to onboard: %s", string(body))
	}

	onboardResponse := &harvester.OnboardHarvesterResponse{}
	if err := json.Unmarshal(body, onboardResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	jwtToken := onboardResponse.Token
	if jwtToken == "" {
		return fmt.Errorf("empty JWT token in onboard response")
	}
	c.jwtStore.setToken(jwtToken)

	c.logger.Info("Connected to Galadriel Server")

	return nil
}

func (c *client) getNewJWTToken(ctx context.Context) error {
	resp, err := c.client.GetNewJWTToken(ctx, c.trustDomain.String())
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	jwtResponse := &harvester.GetJwtResponse{}
	if err := json.Unmarshal(body, jwtResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	jwtToken := jwtResponse.Token
	if jwtToken == "" {
		return fmt.Errorf("JWT token could not be renewed")
	}

	c.logger.Info("JWT token updated")
	c.jwtStore.setToken(jwtToken)

	return nil
}

func createTLSClient(trustBundlePath string) (*http.Client, error) {
	caCert, err := os.ReadFile(trustBundlePath)
	if err != nil {
		return nil, fmt.Errorf("createTLSClient: failed to read trust bundle: %w", err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, fmt.Errorf("failed to append CA certificates")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    caCertPool,
			ServerName: constants.GaladrielServerName,
		},
	}

	return &http.Client{
		Transport: transport,
	}, nil
}

// createEmptyTokenFile creates an empty token file
func (p *jwtStore) createEmptyTokenFile() error {
	// Create the file and close it immediately to create an empty file
	file, err := os.Create(p.tokenFilePath)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}

// createJWTTokenReqEditor creates a request editor function that adds the JWT token to the request's Authorization header
func createJWTTokenReqEditor(jp *jwtStore) harvester.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if req.URL.Path == onboardPath {
			return nil
		}

		token := jp.getToken()
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", constants.JSONContentType)
		return nil
	}
}

func (c *client) startJWTTokenRotation(ctx context.Context) {
	c.logger.Info("Started JWT token rotator")

	ticker := time.NewTicker(jwtRotationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.logger.Debug("Requesting a new JWT token from Galadriel Server")
			if err := c.getNewJWTToken(ctx); err != nil {
				c.logger.Errorf("Error getting new JWT token: %v", err)
			}
		case <-ctx.Done():
			c.logger.Info("JWT token rotator stopped")
			return
		}
	}
}

func createEntityBundle(trustDomainName string, b *harvester.BundlesUpdatesItem) (*entity.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trust domain: %v", err)
	}

	bundleData := []byte(b.TrustBundle)

	digest, err := util.DecodeString(b.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode digest: %v", err)
	}

	signature, err := util.DecodeString(b.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %v", err)
	}

	signingCert, err := util.DecodeString(b.SigningCertificate)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signing certificate: %v", err)
	}

	ret := &entity.Bundle{
		TrustDomainName:    td,
		Data:               bundleData,
		Digest:             digest,
		Signature:          signature,
		SigningCertificate: signingCert,
	}
	return ret, nil
}

func newJwtStore(dataDir, tokenFileName string, logger logrus.FieldLogger) (*jwtStore, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	tokenStoragePath := filepath.Join(dataDir, tokenFileName)
	jp := &jwtStore{
		mu:            sync.RWMutex{},
		jwt:           "",
		tokenFilePath: tokenStoragePath,
		logger:        logger,
	}

	// Load the JWT token from disk storage
	if err := jp.loadToken(); err != nil {
		if os.IsNotExist(err) {
			// Create an empty file if it doesn't exist
			if err := jp.createEmptyTokenFile(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return jp, nil
}

func (p *jwtStore) setToken(jwt string) {
	p.mu.Lock()

	// Sanitize token removing leading and trailing spaces and quotes
	token := strings.TrimSpace(jwt)
	token = strings.ReplaceAll(token, "\"", "")
	p.jwt = token

	// Release the read lock before saving the token to disk
	p.mu.Unlock()

	// Save the token to disk
	if err := p.saveToken(); err != nil {
		p.logger.Errorf("Failed to save JWT token to disk: %v", err)
	}
}

func (p *jwtStore) getToken() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.jwt
}

// loadToken loads the JWT token from the disk storage
func (p *jwtStore) loadToken() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Read the token from the disk storage file
	tokenBytes, err := os.ReadFile(p.tokenFilePath)
	if err != nil {
		return err
	}

	// Set the JWT token
	p.jwt = string(tokenBytes)

	return nil
}

// saveToken saves the JWT token to disk storage
func (p *jwtStore) saveToken() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return diskutil.AtomicWritePrivateFile(p.tokenFilePath, []byte(p.jwt))
}
