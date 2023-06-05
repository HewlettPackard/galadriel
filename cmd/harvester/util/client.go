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
	"github.com/HewlettPackard/galadriel/pkg/harvester/api/admin"
	"github.com/google/uuid"
)

const (
	errFailedRequest          = "failed to send request: %v"
	errUnmarshalRelationships = "failed to unmarshal relationships: %v"
)

// HarvesterAPIClient represents an API client for the Harvester.
type HarvesterAPIClient interface {
	GetRelationships(context.Context, api.ConsentStatus) ([]*entity.Relationship, error)
	UpdateRelationship(context.Context, uuid.UUID, api.ConsentStatus) (*entity.Relationship, error)
}

type harvesterAPIClient struct {
	client *admin.Client
}

// NewUDSClient creates a Harvester API client that connects to the Harvester
// using a Unix Domain Socket (UDS) specified by the socketPath parameter.
func NewUDSClient(socketPath string, httpClient *http.Client) (HarvesterAPIClient, error) {
	clientOpt := admin.WithBaseURL(cli.LocalhostURL)

	client, err := admin.NewClient(socketPath, clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate the Admin Client: %v", err)
	}

	if httpClient == nil {
		httpClient = httputil.NewUDSHTTPClient(socketPath)
	}

	client.Client = httpClient

	return &harvesterAPIClient{client: client}, nil
}

func (h harvesterAPIClient) GetRelationships(ctx context.Context, status api.ConsentStatus) ([]*entity.Relationship, error) {
	payload := &admin.GetRelationshipsParams{ConsentStatus: &status}

	res, err := h.client.GetRelationships(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf(errFailedRequest, err)
	}
	defer res.Body.Close()

	body, err := httputil.ReadResponse(res)
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

	body, err := httputil.ReadResponse(res)
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
