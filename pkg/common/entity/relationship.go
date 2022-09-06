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
	ID                          uuid.UUID
	Status                      string
	SourceMemberID              uuid.UUID
	TargetMemberID              uuid.UUID
	TargetTrustDomain           spiffeid.TrustDomain
	TargetBundleEndpointURL     *url.URL
	TargetBundleEndpointProfile BundleEndpointProfile
	TargetEndpointTrustBundle   *spiffebundle.Bundle
	TargetEndpointSPIFFEID      spiffeid.ID
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}
