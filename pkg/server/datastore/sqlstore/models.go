package sqlstore

import (
	"time"
)

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Organization struct {
	Model
	Name    string `json:"name" gorm:"uniqueIndex"`
	Contact string
	Bridges []Bridge `gorm:"constraint:OnDelete:CASCADE;"`
}

type Bridge struct {
	Model
	OrganizationID uint
	Description    string       `gorm:"uniqueIndex"`
	Memberships    []Membership `gorm:"constraint:OnDelete:CASCADE;"`
}
type Member struct {
	Model
	SpiffeID      string
	Description   string `gorm:"uniqueIndex"`
	Memberships   []Membership
	Relationships []Relationship `gorm:"constraint:OnDelete:CASCADE;"`
	TrustBundles  []TrustBundle  `gorm:"constraint:OnDelete:CASCADE;"`
}
type Membership struct {
	Model
	JoinToken string `gorm:"uniqueIndex"`
	MemberID  uint
	member    Member `gorm:"foreignKey:MemberID;references:ID"`
	BridgeID  uint
	bridge    Bridge `gorm:"foreignKey:BridgeID;references:ID"`
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
