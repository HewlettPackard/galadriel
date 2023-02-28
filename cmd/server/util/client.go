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
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// URL pattern to make http calls on local Unix domain socket,
// the Host is required for the URL, but it's not relevant

const (
	localURL    = "http://local/%s"
	contentType = "application/json"
)

var (
	createTrustDomainURL  = fmt.Sprintf(localURL, "createTrustDomain")
	listTrustDomainsURL   = fmt.Sprintf(localURL, "listTrustDomains")
	createRelationshipURL = fmt.Sprintf(localURL, "createRelationship")
	listRelationshipsURL  = fmt.Sprintf(localURL, "listRelationships")
	generateTokenURL      = fmt.Sprintf(localURL, "generateToken")
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateTrustDomain(m *entity.TrustDomain) error
	ListTrustDomains() ([]*entity.TrustDomain, error)
	CreateRelationship(r *entity.Relationship) error
	ListRelationships() ([]*entity.Relationship, error)
	GenerateJoinToken(trustDomain spiffeid.TrustDomain) (*entity.JoinToken, error)
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

	return serverClient{client: c}
}

type serverClient struct {
	client *http.Client
}

func (c serverClient) CreateTrustDomain(m *entity.TrustDomain) error {
	trustDomainBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	r, err := c.client.Post(createTrustDomainURL, contentType, bytes.NewReader(trustDomainBytes))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if r.StatusCode != 200 {
		return errors.New(string(body))
	}

	return nil
}

func (c serverClient) ListTrustDomains() ([]*entity.TrustDomain, error) {
	r, err := c.client.Get(listTrustDomainsURL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		return nil, errors.New(string(b))
	}

	var trustDomains []*entity.TrustDomain
	if err = json.Unmarshal(b, &trustDomains); err != nil {
		return nil, err
	}

	return trustDomains, nil
}

func (c serverClient) CreateRelationship(rel *entity.Relationship) error {
	relBytes, err := json.Marshal(rel)
	if err != nil {
		return err
	}

	r, err := c.client.Post(createRelationshipURL, contentType, bytes.NewReader(relBytes))
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if r.StatusCode != 200 {
		return errors.New(string(b))
	}

	return nil
}

func (c serverClient) ListRelationships() ([]*entity.Relationship, error) {
	r, err := c.client.Get(listRelationshipsURL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		return nil, errors.New(string(b))
	}

	var rels []*entity.Relationship
	if err = json.Unmarshal(b, &rels); err != nil {
		return nil, err
	}

	return rels, nil
}

func (c serverClient) GenerateJoinToken(td spiffeid.TrustDomain) (*entity.JoinToken, error) {
	b, err := json.Marshal(entity.TrustDomain{Name: td})
	if err != nil {
		return nil, err
	}

	r, err := c.client.Post(generateTokenURL, contentType, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var createdToken entity.JoinToken
	if err = json.Unmarshal(body, &createdToken); err != nil {
		if len(body) == 0 {
			return nil, errors.New("failed to generate token")
		}

		return nil, errors.New(string(body))
	}

	return &createdToken, nil
}
