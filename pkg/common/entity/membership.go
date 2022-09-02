package entity

import "github.com/google/uuid"

type Membership struct {
	FederationGroupId uint
	ID                uuid.UUID
	SpireServerId     uint
	Status            string
}
