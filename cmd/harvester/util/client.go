package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api/admin"
	"io"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
)

const (
	baseURL                   = "http://localhost/"
	errFailedRequest          = "failed to send request: %v"
	errReadResponseBody       = "failed to read response body: %v"
	errUnmarshalRelationships = "failed to unmarshal relationships: %v"
)

// HarvesterAPIClient represents an API client for the Harvester.
type HarvesterAPIClient interface {
	GetRelationships(context.Context, api.ConsentStatus) ([]*entity.Relationship, error)
	UpdateRelationship(context.Context, uuid.UUID, api.ConsentStatus) (*entity.Relationship, error)
}

type ErrorMessage struct {
	Message string
}

type harvesterAPIClient struct {
	client *admin.Client
}

// NewUDSClient creates a Harvester API client that connects to the Harvester
// using a Unix Domain Socket (UDS) specified by the socketPath parameter.
func NewUDSClient(socketPath string) (HarvesterAPIClient, error) {
	clientOpt := admin.WithBaseURL(baseURL)

	client, err := admin.NewClient(socketPath, clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate the Admin Client: %v", err)
	}

	t := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		}}

	client.Client = &http.Client{Transport: t}

	return &harvesterAPIClient{client: client}, nil
}

func (h harvesterAPIClient) GetRelationships(ctx context.Context, status api.ConsentStatus) ([]*entity.Relationship, error) {
	payload := &admin.GetRelationshipsParams{ConsentStatus: &status}

	res, err := h.client.GetRelationships(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readResponse(res)
	if err != nil {
		return nil, err
	}

	var relationships []*api.Relationship
	if err := json.Unmarshal(body, &relationships); err != nil {
		return nil, fmt.Errorf(errUnmarshalRelationships, err)
	}

	rels := make([]*entity.Relationship, 0, len(relationships))
	for i, r := range relationships {
		rel, err := r.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert relationship %d: %v", i, err)
		}
		rels = append(rels, rel)
	}

	return rels, nil
}

func (h harvesterAPIClient) UpdateRelationship(ctx context.Context, relationshipID uuid.UUID, status api.ConsentStatus) (*entity.Relationship, error) {
	payload := admin.PatchRelationshipRequest{ConsentStatus: status}

	res, err := h.client.PatchRelationship(ctx, relationshipID, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := readResponse(res)
	if err != nil {
		return nil, err
	}

	var relationship api.Relationship
	if err := json.Unmarshal(body, &relationship); err != nil {
		return nil, fmt.Errorf(errUnmarshalRelationships, err)
	}

	rel, err := relationship.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert relationship: %v", err)
	}

	return rel, nil
}

func readResponse(res *http.Response) ([]byte, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(errReadResponseBody, err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		var errorMsg ErrorMessage
		if err := json.Unmarshal(body, &errorMsg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error message: %v", err)
		}

		return nil, fmt.Errorf(errorMsg.Message)
	}

	return body, nil
}
