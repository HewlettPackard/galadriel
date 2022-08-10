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
	OrganizationID     uint
	Description        string `gorm:"uniqueIndex"`
	Status             bool
	NestedFedIndicator bool
	Memberships        []Membership `gorm:"constraint:OnDelete:CASCADE;"`
}
type Member struct {
	Model
	SpiffeID          string
	Description       string `gorm:"uniqueIndex"`
	Active            bool
	DiscoverableinDir bool
	AllowDiscovery    bool
	Contact           string
	EndpointURL       string //Type string for now. Maybe changed later on
	SPIREServerInfo   string //Type string for now. Maybe changed later on
	PermissiveMode    bool
	Memberships       []Membership
	Relationships     []Relationship `gorm:"constraint:OnDelete:CASCADE;"`
	TrustBundles      []TrustBundle  `gorm:"constraint:OnDelete:CASCADE;"`
}
type Membership struct {
	Model
	JoinToken     string `gorm:"uniqueIndex"`
	MemberID      uint
	member        Member `gorm:"foreignKey:MemberID;references:ID"`
	BridgeID      uint
	bridge        Bridge `gorm:"foreignKey:BridgeID;references:ID"`
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
