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
	Name    string `json:"name" gorm:"unique_index"`
	Bridges []Bridge
}

type Bridge struct {
	Model
	OrganizationID uint   //Implicit Foreign Key
	Description    string `gorm:"unique_index"`
	Active         bool
	Members        []Member
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
	Memberships       []Membership
	Relationships     []Relationship
	TrustBundles      []TrustBundle
}
type Membership struct {
	Model
	MemberID      uint   //Implicit Foreign Key
	JoinToken     string `gorm:"unique_index"`
	MemberConsent bool
	TTL           uint
	//BridgeID is implicit, as there is a 1:n relationship between Member and Bridge
}

type Relationship struct {
	// Defines the Relationship between two Members
	Model
	MemberID            uint // Implicit Foreign Key and also the SourceID for the relationship
	TargetMemberID      uint
	SourceMemberConsent string
	TargetMemberConsent string
	Status              string
	RelationshipType    string
	TTL                 uint
}

type TrustBundle struct {
	Model
	MemberID    uint //Implicit Foreign Key
	TrustBundle string
}
