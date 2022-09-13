package sqlstore

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BeforeCreate adds a gorm hook to insert the uuid before creation
func (m *Model) BeforeCreate(db *gorm.DB) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	m.ID = id
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
	Status        string
	AccessTokens  []AccessToken
	Memberships   []Membership   `gorm:"constraint:OnDelete:CASCADE"`
	Relationships []Relationship `gorm:"constraint:OnDelete:CASCADE;foreignKey:SourceMemberID"`
	TrustBundle   string
}

type AccessToken struct {
	Model

	MemberID uuid.UUID `gorm:"type:uuid"`
	Value    string
}

type FederationGroup struct {
	Model
}

type Membership struct {
	Model

	MemberID          uuid.UUID `gorm:"type:uuid"`
	FederationGroupID uuid.UUID `gorm:"type:uuid"`
}

type Relationship struct {
	Model

	SourceMemberID uuid.UUID `gorm:"type:uuid"`
	TargetMemberID uuid.UUID `gorm:"type:uuid"`
	Status         string
}

type Migration struct {
	Model
	// Database version
	Version int
}
