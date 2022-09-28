package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// URL pattern to make http calls on local Unix domain socket,
// the Host is required for the URL, but it's not relevant

const (
	localURL    = "http://local"
	contentType = "application/json"

	createMemberURL       = localURL + admin.CreateMemberPath
	createRelationshipURL = localURL + admin.CreateRelationshipPath
	generateTokenURL      = localURL + admin.GenerateTokenPath
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateMember(m *common.Member) error
	CreateRelationship(r *common.Relationship) error
	GenerateAccessToken(trustDomain spiffeid.TrustDomain) (*common.AccessToken, error)
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

func (c serverClient) CreateMember(m *common.Member) error {
	memberBytes, err := json.Marshal(m)

	if err != nil {
		return err
	}

	r, err := c.client.Post(createMemberURL, contentType, bytes.NewReader(memberBytes))
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

func (c serverClient) CreateRelationship(rel *common.Relationship) error {
	relBytes, err := json.Marshal(rel)
	if err != nil {
		return err
	}

	r, err := c.client.Post(createRelationshipURL, contentType, bytes.NewReader(relBytes))
	if err != nil {
		return err
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

func (c serverClient) GenerateAccessToken(td spiffeid.TrustDomain) (*common.AccessToken, error) {
	b, err := json.Marshal(common.Member{TrustDomain: td})
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

	var createdToken common.AccessToken
	if err = json.Unmarshal(body, &createdToken); err != nil {
		if len(body) == 0 {
			return nil, errors.New("failed to generate token")
		}

		return nil, errors.New(string(body))
	}

	return &createdToken, nil
}
