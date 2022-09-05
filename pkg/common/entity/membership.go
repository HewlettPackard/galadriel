package entity

import (
	"time"

	"github.com/google/uuid"
)

type Membership struct {
	ID                uuid.UUID //SourceMemberID
	TargetMemberID    uuid.UUID
	FederationGroupID uuid.UUID
	MemberID          uuid.UUID
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
