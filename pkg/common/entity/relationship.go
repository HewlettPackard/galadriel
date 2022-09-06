package entity

import (
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type BundleEndpointProfile interface {
	Name() string
}

type Relationship struct {
	ID                    uuid.UUID
	Status                string
	SourceMemberID        uuid.UUID
	TargetMemberID        uuid.UUID
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     *url.URL
	BundleEndpointProfile BundleEndpointProfile
	TrustDomainBundle     *spiffebundle.Bundle
	EndpointSPIFFEID      spiffeid.ID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
