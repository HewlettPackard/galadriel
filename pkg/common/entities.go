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
	Token    string
	Expiry   time.Time
	MemberID uuid.UUID
}

type Member struct {
	ID uuid.UUID

	Name        string
	TrustDomain spiffeid.TrustDomain
}
