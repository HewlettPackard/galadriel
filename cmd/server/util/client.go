package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// URL pattern to make http calls on local Unix domain socket,
// the Host is required for the URL, but it's not relevant

const (
	localURL    = "http://local/%s"
	contentType = "application/json"
)

var (
	trustDomainByNameURL = fmt.Sprintf(localURL, "trust-domain/%s")
	getJoinTokenURL      = fmt.Sprintf(localURL, "trust-domain/%s/join-token")
	trustDomainURL       = fmt.Sprintf(localURL, "trust-domain")
	relationshipsURL     = fmt.Sprintf(localURL, "relationships")
	relationshipsByIDURL = fmt.Sprintf(localURL, "relationships/%s")
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateTrustDomain(ctx context.Context, trustDomain *entity.TrustDomain) (*entity.TrustDomain, error)
	GetTrustDomainByName(ctx context.Context, trustDomainName spiffeid.TrustDomain) (*entity.TrustDomain, error)
	UpdateTrustDomainByName(ctx context.Context, trustDomainName spiffeid.TrustDomain) (*entity.TrustDomain, error)
	CreateRelationship(ctx context.Context, r *entity.Relationship) (*entity.Relationship, error)
	GetRelationshipByID(ctx context.Context, id uuid.UUID) (*entity.Relationship, error)
	GetRelationships(ctx context.Context, consentStatus string, trustDomain string) (*entity.Relationship, error)
	GetJoinToken(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.JoinToken, error)
}

// TODO: improve this adding options for the transport, dialcontext, and http.Client.
func NewServerClient(socketPath string) ServerLocalClient {
	t := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		}}
	c := &http.Client{
		Transport: t,
	}

	return &serverClient{client: c}
}

type serverClient struct {
	client *http.Client
}

func (c *serverClient) GetTrustDomainByName(ctx context.Context, trustDomainName spiffeid.TrustDomain) (*entity.TrustDomain, error) {
	trustDomainNameBytes, err := trustDomainName.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trust domain: %v", err)
	}

	getTrustDomainByNameURL := fmt.Sprintf(trustDomainByNameURL, trustDomainName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getTrustDomainByNameURL, bytes.NewReader(trustDomainNameBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
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

func (c *serverClient) UpdateTrustDomainByName(ctx context.Context, trustDomainName spiffeid.TrustDomain) (*entity.TrustDomain, error) {
	trustDomainNameBytes, err := trustDomainName.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trust domain: %v", err)
	}

	updateTrustDomainByNameURL := fmt.Sprintf(trustDomainByNameURL, trustDomainName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, updateTrustDomainByNameURL, bytes.NewReader(trustDomainNameBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(req.Body)
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
	trustDomainBytes, err := json.Marshal(trustDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trust domain: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, trustDomainURL, bytes.NewReader(trustDomainBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(req.Body)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, relationshipsURL, bytes.NewReader(relBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var relationships *entity.Relationship
	if err = json.Unmarshal(body, &relationships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trust domain: %v", err)
	}

	return relationships, nil
}
func (c *serverClient) GetRelationships(ctx context.Context, consentStatus string, trustDomain string) (*entity.Relationship, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, relationshipsURL, bytes.NewReader())
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(req.Body)
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

func (c *serverClient) GetRelationshipByID(ctx context.Context, id uuid.UUID) (*entity.Relationship, error) {
	relationshipsByIDURL = fmt.Sprintf(relationshipsByIDURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, relationshipsByIDURL, bytes.NewReader(id[:]))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(req.Body)
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

func (c *serverClient) GetJoinToken(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.JoinToken, error) {
	trustDomainBytes, err := trustDomain.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trust domain: %v", err)
	}

	getJoinTokenURL = fmt.Sprintf(getJoinTokenURL, trustDomain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getJoinTokenURL, bytes.NewReader(trustDomainBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.client.Do(req)
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

// func (c clientServer) GenerateJoinToken(td spiffeid.TrustDomain) (string, error) {
// 	joinTokenURL := fmt.Sprintf(joinTokenURL, td)
// 	r, err := c.client.Get(joinTokenURL)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer r.Body.Close()

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(body), nil
// }
