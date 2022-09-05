package sqlstore

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Added a gorm hook to insert the uuid before creation
func (m *Model) BeforeCreate(db *gorm.DB) error {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	m.ID = uuid
	return nil
}

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
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
	MemberID  uuid.UUID
	TTL       uint
}

type Relationship struct {
	Model
	MemberID         uuid.UUID // Implicit Foreign Key and also the SourceID for the relationship
	TargetMemberID   uuid.UUID `gorm:"uniqueIndex"`
	Status           string
	RelationshipType string
	TTL              uint
}

type TrustBundle struct {
	Model
	MemberID    uuid.UUID //Implicit Foreign Key
	TrustBundle string
}

type Migration struct {
	Model
	// Database version
	Version int
}
