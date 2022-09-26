package common

import (
	"time"

	"github.com/google/uuid"
)

type Relationship struct {
	ID uuid.UUID

	TrustDomainA string
	TrustDomainB string
}

type AccessToken struct {
	Token    string
	Expiry   time.Time
	MemberID uuid.UUID
}

type Member struct {
	ID uuid.UUID

	Name        string
	TrustDomain string
}
