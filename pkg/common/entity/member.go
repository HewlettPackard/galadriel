package entity

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	MemberID    uuid.UUID
	Description string
	Status      string
	TrustDomain string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
