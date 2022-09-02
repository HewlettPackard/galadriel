package entity

import (
	"time"

	"github.com/google/uuid"
)

type FederationGroup struct {
	FederationGroupID uuid.UUID
	Name              string
	Orgid             uint
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
