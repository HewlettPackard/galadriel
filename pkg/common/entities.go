package common

import (
	"github.com/google/uuid"
	"time"
)

type Relationship struct {
	MemberA uuid.UUID
	MemberB uuid.UUID
}

type AccessToken struct {
	Token  string
	Expiry time.Time
}

type Member struct {
	ID uuid.UUID

	Name   string
	Tokens []AccessToken
}
