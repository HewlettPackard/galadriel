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
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// URL pattern to make http calls on local Unix domain socket,
// the Host is required for the URL, but it's not relevant

const (
	localURL    = "http://local/%s"
	contentType = "application/json"
)

var (
	createMemberURL       = fmt.Sprintf(localURL, "createMember")
	listMembersURL        = fmt.Sprintf(localURL, "listMembers")
	createRelationshipURL = fmt.Sprintf(localURL, "createRelationship")
	listRelationshipsURL  = fmt.Sprintf(localURL, "listRelationships")
	generateTokenURL      = fmt.Sprintf(localURL, "generateToken")
)

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	CreateMember(m *common.Member) error
	ListMembers() ([]*common.Member, error)
	CreateRelationship(r *common.Relationship) error
	ListRelationships() ([]*common.Relationship, error)
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

func (c serverClient) ListMembers() ([]*common.Member, error) {
	r, err := c.client.Get(listMembersURL)
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

	var members []*common.Member
	if err = json.Unmarshal(b, &members); err != nil {
		return nil, err
	}

	return members, nil
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

func (c serverClient) ListRelationships() ([]*common.Relationship, error) {
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

	var rels []*common.Relationship
	if err = json.Unmarshal(b, &rels); err != nil {
		return nil, err
	}

	return rels, nil
}

func (c serverClient) GenerateAccessToken(td spiffeid.TrustDomain) (*common.AccessToken, error) {
	b, err := json.Marshal(common.Member{TrustBundle: common.TrustBundle{TrustDomain: td}})
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
