package endpoints

import (
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var (
	tdA           = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-a.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdB           = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-b.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdC           = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-c.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	pendingRelAB  = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	pendingRelAC  = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	approvedRelAB = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	approvedRelAC = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	approvedRelBC = &entity.Relationship{TrustDomainAID: tdB.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
)
