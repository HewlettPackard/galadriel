package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type ConsentStatus string

const (
	ConsentStatusApproved ConsentStatus = "approved"
	ConsentStatusDenied   ConsentStatus = "denied"
	ConsentStatusPending  ConsentStatus = "pending"
)

type TrustDomain struct {
	ID          uuid.NullUUID
	Name        spiffeid.TrustDomain
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Relationship struct {
	ID                  uuid.NullUUID
	TrustDomainAID      uuid.UUID
	TrustDomainBID      uuid.UUID
	TrustDomainAName    spiffeid.TrustDomain
	TrustDomainBName    spiffeid.TrustDomain
	TrustDomainAConsent ConsentStatus
	TrustDomainBConsent ConsentStatus
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type JoinToken struct {
	ID              uuid.NullUUID
	Token           string
	Used            bool
	TrustDomainID   uuid.UUID
	TrustDomainName spiffeid.TrustDomain
	ExpiresAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Bundle represents a SPIFFE Trust bundle along with its digest.
type Bundle struct {
	ID                 uuid.NullUUID
	Data               []byte // Raw bundle data.
	Digest             []byte // SHA-256 digest of the bundle data.
	Signature          []byte // Raw signature of the bundle data.
	SigningCertificate []byte
	TrustDomainID      uuid.UUID
	TrustDomainName    spiffeid.TrustDomain
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
