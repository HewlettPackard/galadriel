package sqlstore

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Member struct {
	Model
	Description   string
	TrustDomain   string
	Memberships   []Membership   `gorm:"constraint:OnDelete:CASCADE;"`
	Relationships []Relationship `gorm:"constraint:OnDelete:CASCADE;"`
	TrustBundles  []TrustBundle  `gorm:"constraint:OnDelete:CASCADE;"`
}

type Membership struct {
	Model
	JoinToken string `gorm:"uniqueIndex"`
	MemberID  uint
	TTL       uint
}

type Relationship struct {
	Model
	MemberID         uint // Implicit Foreign Key and also the SourceID for the relationship
	TargetMemberID   uint
	Status           string
	RelationshipType string
	TTL              uint
}

type TrustBundle struct {
	Model
	MemberID    uint //Implicit Foreign Key
	TrustBundle string
}

type Migration struct {
	Model
	// Database version
	Version int
}
