package entity

import "github.com/google/uuid"

type FederationGroup struct {
	ID     uuid.UUID
	Name   string
	Orgid  uint
	Status string
}
