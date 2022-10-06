package common

import (
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type Relationship struct {
	ID uuid.UUID

	TrustDomainA spiffeid.TrustDomain
	TrustDomainB spiffeid.TrustDomain
}

type AccessToken struct {
	MemberID uuid.UUID

	Token  string
	Expiry time.Time
}

type Member struct {
	ID uuid.UUID

	Name            string
	TrustDomain     spiffeid.TrustDomain
	TrustBundle     []byte
	TrustBundleHash []byte
}
