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
	Status        string
	Memberships   []Membership   `gorm:"constraint:OnDelete:CASCADE"`
	Relationships []Relationship `gorm:"constraint:OnDelete:CASCADE;foreignKey:SourceMemberID"`
	TrustBundle   string
}

type Membership struct {
	Model
	JoinToken string    `gorm:"unique_index"`
	MemberID  uuid.UUID `gorm:"type:uuid"`
	TTL       uint
}

type Relationship struct {
	Model
	SourceMemberID              uuid.UUID `gorm:"type:uuid"`
	TargetMemberID              uuid.UUID `gorm:"type:uuid"`
	Status                      string
	TargetTrustDomain           string
	TargetBundleEndpointURL     string
	TargetBundleEndpointProfile string
	TargetEndpointTrustBundle   string
	TargetEndpointSPIFFEID      string
}

type Migration struct {
	Model
	// Database version
	Version int
}
