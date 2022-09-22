package common

import (
	"time"

	"github.com/google/uuid"
)

type Relationship struct {
	ID uuid.UUID

	MemberA uuid.UUID
	MemberB uuid.UUID
}

type AccessToken struct {
	Token  string
	Expiry time.Time
}

type Member struct {
	ID uuid.UUID

	Name        string
	TrustDomain string
	Tokens      []AccessToken
}
