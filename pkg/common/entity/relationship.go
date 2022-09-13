package entity

import (
	"time"

	"github.com/google/uuid"
)

type Relationship struct {
	ID          uuid.UUID
	Status      string
	MemberOneID uuid.UUID
	MemberTwoID uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
