// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package datastore

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

type ConsentStatus string

const (
	ConsentStatusAccepted ConsentStatus = "accepted"
	ConsentStatusDenied   ConsentStatus = "denied"
	ConsentStatusPending  ConsentStatus = "pending"
)

func (e *ConsentStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ConsentStatus(s)
	case string:
		*e = ConsentStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ConsentStatus: %T", src)
	}
	return nil
}

type NullConsentStatus struct {
	ConsentStatus ConsentStatus
	Valid         bool // Valid is true if ConsentStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullConsentStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ConsentStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ConsentStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullConsentStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ConsentStatus), nil
}

type Bundle struct {
	ID                 pgtype.UUID
	TrustDomainID      pgtype.UUID
	Data               []byte
	Signature          []byte
	SignatureAlgorithm sql.NullString
	SigningCertificate []byte
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type JoinToken struct {
	ID            pgtype.UUID
	TrustDomainID pgtype.UUID
	Token         string
	Used          bool
	ExpiresAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Relationship struct {
	ID                  pgtype.UUID
	TrustDomainAID      pgtype.UUID
	TrustDomainBID      pgtype.UUID
	TrustDomainAConsent ConsentStatus
	TrustDomainBConsent ConsentStatus
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type TrustDomain struct {
	ID                pgtype.UUID
	Name              string
	Description       sql.NullString
	HarvesterSpiffeID sql.NullString
	OnboardingBundle  []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
