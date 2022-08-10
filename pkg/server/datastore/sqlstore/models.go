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
	Name    string   `json:"name" gorm:"unique_index"`
	Bridges []Bridge `gorm:"constraint:OnDelete:CASCADE;"`
}

type Bridge struct {
	Model
	OrganizationID uint
	Description    string `gorm:"unique_index"`
	Active         bool
	Members        []Member `gorm:"constraint:OnDelete:CASCADE;"`
}
type Member struct {
	Model
	BridgeID          uint //Implicit Foreign Key
	SpiffeID          string
	Description       string `gorm:"unique_index"`
	Active            bool
	DiscoverableinDir bool
	AllowDiscovery    bool
	Contact           string
	EndpointURL       string //Type string for now. Maybe changed later on
	SPIREServerInfo   string //Type string for now. Maybe changed later on
	PermissiveMode    bool
	Memberships       []Membership   `gorm:"constraint:OnDelete:CASCADE;"`
	Relationships     []Relationship `gorm:"constraint:OnDelete:CASCADE;"`
	TrustBundles      []TrustBundle  `gorm:"constraint:OnDelete:CASCADE;"`
}
type Membership struct {
	Model
	MemberID      uint   //Implicit Foreign Key
	JoinToken     string `gorm:"unique_index"`
	MemberConsent bool
	TTL           uint
}

type Relationship struct {
	Model
	MemberID            uint // Implicit Foreign Key and also the SourceID for the relationship
	TargetMemberID      uint
	SourceMemberConsent bool
	TargetMemberConsent bool
	Status              string
	RelationshipType    string
	TTL                 uint
}

type TrustBundle struct {
	Model
	MemberID    uint //Implicit Foreign Key
	TrustBundle string
}
