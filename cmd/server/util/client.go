package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/google/uuid"
)

const (
	baseURL                   = "http://localhost/"
	errFailedRequest          = "failed to send request: %v"
	errReadResponseBody       = "failed to read response body: %v"
	errStatusCodeMsg          = "request returned an error code %d: \n%s"
	errUnmarshalRelationships = "failed to unmarshal relationships: %v"
	errUnmarshalTrustDomains  = "failed to unmarshal trust domain: %v"
	errUnmarshalJoinToken     = "failed to unmarshal join token: %v"
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateTrustDomain(ctx context.Context, trustDomain api.TrustDomainName) (*entity.TrustDomain, error)
	GetTrustDomainByName(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error)
	UpdateTrustDomainByName(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error)
	CreateRelationship(ctx context.Context, r *entity.Relationship) (*entity.Relationship, error)
	GetRelationshipByID(ctx context.Context, relID uuid.UUID) (*entity.Relationship, error)
	GetRelationships(ctx context.Context, consentStatus api.ConsentStatus, trustDomainName api.TrustDomainName) (*entity.Relationship, error)
	GetJoinToken(ctx context.Context, trustDomain api.TrustDomainName) (*entity.JoinToken, error)
}

func NewServerClient(socketPath string) (ServerLocalClient, error) {
	clientOpt := admin.WithBaseURL(baseURL)

	client, err := admin.NewClient(socketPath, clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate the Admin Client: %q", err)
	}

	t := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		}}

	client.Client = &http.Client{Transport: t}

	return &serverClient{client: client}, nil
}

type serverClient struct {
	client *admin.Client
}

func (c *serverClient) GetTrustDomainByName(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	res, err := c.client.GetTrustDomainByName(ctx, trustDomainName)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
	if err != nil {
		return nil, err
	}

	trustDomain, err := unmarshalTrustDomain(body)
	if err != nil {
		return nil, err
	}

	return trustDomain, nil
}

func (c *serverClient) UpdateTrustDomainByName(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	payload := api.TrustDomain{Name: trustDomainName}
	res, err := c.client.PutTrustDomainByName(ctx, trustDomainName, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
	if err != nil {
		return nil, err
	}

	trustDomain, err := unmarshalTrustDomain(body)
	if err != nil {
		return nil, err
	}

	return trustDomain, nil
}

func (c *serverClient) CreateTrustDomain(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	payload := admin.PutTrustDomainJSONRequestBody{Name: trustDomainName}

	res, err := c.client.PutTrustDomain(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
	if err != nil {
		return nil, err
	}

	trustDomain, err := unmarshalTrustDomain(body)
	if err != nil {
		return nil, err
	}

	return trustDomain, nil
}

func (c *serverClient) CreateRelationship(ctx context.Context, rel *entity.Relationship) (*entity.Relationship, error) {
	payload := admin.PutRelationshipJSONRequestBody{TrustDomainAName: rel.TrustDomainAName.String(), TrustDomainBName: rel.TrustDomainBName.String()}
	res, err := c.client.PutRelationship(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
	if err != nil {
		return nil, err
	}

	relationship, err := unmarshalRelationship(body)
	if err != nil {
		return nil, err
	}

	return relationship, nil
}
func (c *serverClient) GetRelationships(ctx context.Context, consentStatus api.ConsentStatus, trustDomainName api.TrustDomainName) (*entity.Relationship, error) {
	payload := &admin.GetRelationshipsParams{Status: (*api.ConsentStatus)(&consentStatus), TrustDomainName: &trustDomainName}

	res, err := c.client.GetRelationships(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
	if err != nil {
		return nil, err
	}

	relationship, err := unmarshalRelationship(body)
	if err != nil {
		return nil, err
	}

	return relationship, nil
}

func (c *serverClient) GetRelationshipByID(ctx context.Context, relID uuid.UUID) (*entity.Relationship, error) {
	res, err := c.client.GetRelationshipByID(ctx, relID)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
	if err != nil {
		return nil, err
	}

	relationship, err := unmarshalRelationship(body)
	if err != nil {
		return nil, err
	}

	return relationship, nil
}

func (c *serverClient) GetJoinToken(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.JoinToken, error) {
	res, err := c.client.GetJoinToken(ctx, trustDomainName)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readBodyAndStatusCode(res)
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

func readBodyAndStatusCode(res *http.Response) ([]byte, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(errReadResponseBody, err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(errStatusCodeMsg, res.StatusCode, body)
	}

	return body, nil
}
