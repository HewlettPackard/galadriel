package entity

import (
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire/proto/spire/common"
)

type BundleEndpointType string

type Relationship struct {
	ID                    uuid.UUID
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     *url.URL
	BundleEndpointProfile BundleEndpointType
	TrustDomainBundle     *common.Bundle
	EndpointSPIFFEID      spiffeid.ID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
