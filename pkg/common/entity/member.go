package entity

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	ID          uuid.UUID
	Description string
	Status      string
	TrustDomain string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
