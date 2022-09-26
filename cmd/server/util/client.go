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

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
)

// URL pattern to make http calls on local Unix domain socket,
// the Host is required for the URL, but it's not relevant

const (
	localURL = "http://local/%s"

	generateTokenPath      = "generateToken"
	createMemberPath       = "createMember"
	createRelationshipPath = "createRelationship"

	contentType = "application/json"
)

var (
	createMemberURL       = fmt.Sprintf(localURL, createMemberPath)
	createRelationshipURL = fmt.Sprintf(localURL, createRelationshipPath)
	generateTokenURL      = fmt.Sprintf(localURL, generateTokenPath)
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateMember(m common.Member) error
	CreateRelationship(r common.Relationship) error
	GenerateAccessToken(trustDomain string) (*datastore.AccessToken, error)
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

func (c serverClient) CreateMember(m common.Member) error {
	if err := validateMember(m); err != nil {
		return err
	}

	memberBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	bytes.NewBuffer(memberBytes)
	_, err = c.client.Post(createMemberURL, contentType, bytes.NewBuffer(memberBytes))
	if err != nil {
		return err
	}

	return nil
}

func (c serverClient) CreateRelationship(rel common.Relationship) error {
	if err := validateRelationship(rel); err != nil {
		return err
	}

	relBytes, err := json.Marshal(rel)
	if err != nil {
		return err
	}

	bytes.NewBuffer(relBytes)
	r, err := c.client.Post(createRelationshipURL, contentType, bytes.NewBuffer(relBytes))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var createdRelationship datastore.Relationship
	if err = json.Unmarshal(b, &createdRelationship); err != nil {
		return err
	}

	return nil
}

func (c serverClient) GenerateAccessToken(trustDomain string) (*datastore.AccessToken, error) {
	b, err := json.Marshal(common.Member{TrustDomain: trustDomain})
	if err != nil {
		return nil, err
	}

	r, err := c.client.Post(generateTokenURL, contentType, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var createdToken datastore.AccessToken
	if err = json.Unmarshal(body, &createdToken); err != nil {
		if len(body) == 0 {
			// TODO: validation based on error codes/status check
			return nil, errors.New("not found")
		}

		return nil, err
	}

	return &createdToken, nil
}

func validateMember(m common.Member) error {
	// TODO: checks
	return nil
}

func validateRelationship(m common.Relationship) error {
	// TODO: checks
	return nil
}
