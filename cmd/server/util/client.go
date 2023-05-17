package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/google/uuid"
)

const (
	jsonContentType = "application/json"
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateTrustDomain(ctx context.Context, trustDomain *entity.TrustDomain) (*entity.TrustDomain, error)
	GetTrustDomainByName(ctx context.Context, trustDomainName string) (*entity.TrustDomain, error)
	UpdateTrustDomainByName(ctx context.Context, trustDomainName string) (*entity.TrustDomain, error)
	CreateRelationship(ctx context.Context, r *entity.Relationship) (*entity.Relationship, error)
	GetRelationshipByID(ctx context.Context, relID uuid.UUID) (*entity.Relationship, error)
	GetRelationships(ctx context.Context, consentStatus string, trustDomain string) (*entity.Relationship, error)
	GetJoinToken(ctx context.Context, trustDomain string) (*entity.JoinToken, error)
}

func NewServerClient(socketPath string) ServerLocalClient {
	return &serverClient{client: &admin.Client{Server: socketPath}}
}

type serverClient struct {
	client *admin.Client
}

func (c *serverClient) GetTrustDomainByName(ctx context.Context, trustDomainName string) (*entity.TrustDomain, error) {
	res, err := c.client.GetTrustDomainByName(ctx, trustDomainName)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var trustDomain *entity.TrustDomain
	if err = json.Unmarshal(body, &trustDomain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trust domain: %v", err)
	}

	return trustDomain, nil
}

func (c *serverClient) UpdateTrustDomainByName(ctx context.Context, trustDomainName string) (*entity.TrustDomain, error) {
	res, err := c.client.PutTrustDomainByNameWithBody(ctx, trustDomainName, jsonContentType, bytes.NewReader([]byte(trustDomainName)))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var trustDomain *entity.TrustDomain
	if err = json.Unmarshal(body, &trustDomain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trust domain: %v", err)
	}

	return trustDomain, nil
}

func (c *serverClient) CreateTrustDomain(ctx context.Context, trustDomain *entity.TrustDomain) (*entity.TrustDomain, error) {
	payload := admin.PutTrustDomainJSONRequestBody{Name: trustDomain.Name.String()}

	res, err := c.client.PutTrustDomain(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var trustDomainRes *entity.TrustDomain
	if err = json.Unmarshal(body, &trustDomainRes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trust domain: %v", err)
	}

	return trustDomain, nil
}

func (c *serverClient) CreateRelationship(ctx context.Context, rel *entity.Relationship) (*entity.Relationship, error) {
	relBytes, err := json.Marshal(rel)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Relationship: %v", err)
	}

	res, err := c.client.PutRelationshipWithBody(ctx, jsonContentType, bytes.NewReader(relBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var relationships *entity.Relationship
	if err = json.Unmarshal(body, &relationships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relationships: %v", err)
	}

	return relationships, nil
}
func (c *serverClient) GetRelationships(ctx context.Context, consentStatus string, trustDomain string) (*entity.Relationship, error) {
	payload := &admin.GetRelationshipsParams{Status: (*api.ConsentStatus)(&consentStatus), TrustDomainName: &trustDomain}

	res, err := c.client.GetRelationships(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var relationship *entity.Relationship
	if err = json.Unmarshal(body, &relationship); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relationship: %v", err)
	}

	return relationship, nil
}

func (c *serverClient) GetRelationshipByID(ctx context.Context, relID uuid.UUID) (*entity.Relationship, error) {
	res, err := c.client.GetRelationshipByID(ctx, relID)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var relationship *entity.Relationship
	if err = json.Unmarshal(body, &relationship); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relationship: %v", err)
	}

	return relationship, nil
}

func (c *serverClient) GetJoinToken(ctx context.Context, trustDomainName string) (*entity.JoinToken, error) {
	res, err := c.client.GetJoinToken(ctx, trustDomainName)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var joinToken *entity.JoinToken
	if err = json.Unmarshal(body, &joinToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal join token: %v", err)
	}

	return joinToken, nil
}
