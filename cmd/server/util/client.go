package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"
	httputil "github.com/HewlettPackard/galadriel/cmd/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/google/uuid"
)

const (
	errFailedRequest          = "failed to send request: %v"
	errUnmarshalRelationships = "failed to unmarshal relationships: %v"
	errUnmarshalTrustDomains  = "failed to unmarshal trust domain: %v"
	errUnmarshalJoinToken     = "failed to unmarshal join token: %v"
)

// GaladrielAPIClient represents an API client for the Galadriel Server.
type GaladrielAPIClient interface {
	CreateTrustDomain(context.Context, api.TrustDomainName) (*entity.TrustDomain, error)
	GetTrustDomainByName(context.Context, api.TrustDomainName) (*entity.TrustDomain, error)
	UpdateTrustDomainByName(context.Context, api.TrustDomainName) (*entity.TrustDomain, error)
	CreateRelationship(context.Context, *entity.Relationship) (*entity.Relationship, error)
	GetRelationshipByID(context.Context, uuid.UUID) (*entity.Relationship, error)
	GetRelationships(context.Context, api.ConsentStatus, api.TrustDomainName) (*entity.Relationship, error)
	GetJoinToken(context.Context, api.TrustDomainName, int32) (*entity.JoinToken, error)
}

type serverAPIClient struct {
	client *admin.Client
}

// NewUDSClient creates a Galadriel API client that connects to the Galadriel Server
// using a Unix Domain Socket (UDS) specified by the socketPath parameter.
func NewUDSClient(socketPath string, httpClient *http.Client) (GaladrielAPIClient, error) {
	clientOpt := admin.WithBaseURL(cli.LocalhostURL)

	client, err := admin.NewClient(socketPath, clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate the Admin Client: %v", err)
	}

	if httpClient == nil {
		httpClient = httputil.NewUDSHTTPClient(socketPath)
	}

	client.Client = httpClient

	return &serverAPIClient{client: client}, nil
}

func (c *serverAPIClient) GetTrustDomainByName(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	res, err := c.client.GetTrustDomainByName(ctx, trustDomainName)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	trustDomain, err := unmarshalTrustDomain(body)
	if err != nil {
		return nil, err
	}

	return trustDomain, nil
}

func (c *serverAPIClient) UpdateTrustDomainByName(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	payload := api.TrustDomain{Name: trustDomainName}
	res, err := c.client.PutTrustDomainByName(ctx, trustDomainName, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	trustDomain, err := unmarshalTrustDomain(body)
	if err != nil {
		return nil, err
	}

	return trustDomain, nil
}

func (c *serverAPIClient) CreateTrustDomain(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	payload := admin.PutTrustDomainJSONRequestBody{Name: trustDomainName}

	res, err := c.client.PutTrustDomain(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	trustDomain, err := unmarshalTrustDomain(body)
	if err != nil {
		return nil, err
	}

	return trustDomain, nil
}

func (c *serverAPIClient) CreateRelationship(ctx context.Context, rel *entity.Relationship) (*entity.Relationship, error) {
	payload := admin.PutRelationshipJSONRequestBody{TrustDomainAName: rel.TrustDomainAName.String(), TrustDomainBName: rel.TrustDomainBName.String()}
	res, err := c.client.PutRelationship(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	relationship, err := unmarshalRelationship(body)
	if err != nil {
		return nil, err
	}

	return relationship, nil
}
func (c *serverAPIClient) GetRelationships(ctx context.Context, consentStatus api.ConsentStatus, trustDomainName api.TrustDomainName) (*entity.Relationship, error) {
	payload := &admin.GetRelationshipsParams{Status: &consentStatus, TrustDomainName: &trustDomainName}

	res, err := c.client.GetRelationships(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	relationship, err := unmarshalRelationship(body)
	if err != nil {
		return nil, err
	}

	return relationship, nil
}

func (c *serverAPIClient) GetRelationshipByID(ctx context.Context, relID uuid.UUID) (*entity.Relationship, error) {
	res, err := c.client.GetRelationshipByID(ctx, relID)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	relationship, err := unmarshalRelationship(body)
	if err != nil {
		return nil, err
	}

	return relationship, nil
}

func (c *serverAPIClient) GetJoinToken(ctx context.Context, trustDomainName api.TrustDomainName, ttl int32) (*entity.JoinToken, error) {
	params := &admin.GetJoinTokenParams{Ttl: ttl}
	res, err := c.client.GetJoinToken(ctx, trustDomainName, params)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
	if err != nil {
		return nil, err
	}

	var joinToken *entity.JoinToken
	if err = json.Unmarshal(body, &joinToken); err != nil {
		return nil, fmt.Errorf(errUnmarshalJoinToken, err)
	}

	return joinToken, nil
}

func unmarshalTrustDomain(body []byte) (*entity.TrustDomain, error) {
	var trustDomain *entity.TrustDomain
	if err := json.Unmarshal(body, &trustDomain); err != nil {
		return nil, fmt.Errorf(errUnmarshalTrustDomains, err)
	}

	return trustDomain, nil
}

func unmarshalRelationship(body []byte) (*entity.Relationship, error) {
	var relationship *entity.Relationship
	if err := json.Unmarshal(body, &relationship); err != nil {
		return nil, fmt.Errorf(errUnmarshalRelationships, err)
	}

	return relationship, nil
}
