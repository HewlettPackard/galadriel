package entity

import (
	"github.com/google/uuid"
)

type Member struct {
	ID          uuid.UUID
	Description string
	Status      string
	TrustDomain string
	TrustBundle string
	Token       []string
}
