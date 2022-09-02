package entity

import "github.com/google/uuid"

type Member struct {
	Description string
	ID          uuid.UUID
	Status      string
	TrustDomain string
}
