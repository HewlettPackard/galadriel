package entity

import (
	"time"

	"github.com/google/uuid"
)

type Membership struct {
	MembershipID      uuid.UUID
	FederationGroupId uint
	SpireServerId     uint
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
