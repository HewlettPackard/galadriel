package sqlstore

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Member struct {
	gorm.Model
	BridgeID        uint
	SpiffeID        string `gorm:"unique_index"`
	Description     string
	Active          bool
	DiscoverableDir bool
	AllowDiscovery  bool
	Contact         string
	EndpointURL     string //Type string for now. Maybe changed later on
	SPIREServerInfo string //Type string for now. Maybe changed later on
	PermissiveMode  bool
}

type Bridge struct {
	gorm.Model
	OrganizationID uint
	Description    string `gorm:"unique_index"`
	Active         bool
	Members        []Member
}

type Organization struct {
	Model
	Name    string `json:"name" gorm:"unique_index"`
	Bridges []Bridge
}
