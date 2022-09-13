package entity

import (
	"time"

	"github.com/google/uuid"
)

type FederationGroup struct {
	ID        uuid.UUID
	Name      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
