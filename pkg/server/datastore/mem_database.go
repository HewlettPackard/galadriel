package datastore

import (
	"context"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type InMemoryDatabase struct {
	mutex  sync.Mutex
	toFail bool

	// Entities
	bundles       map[uuid.UUID]*entity.Bundle
	tokens        map[uuid.UUID]*entity.JoinToken
	trustDomains  map[uuid.UUID]*entity.TrustDomain
	relationships map[uuid.UUID]*entity.Relationship
}

func NewInMemoryDB() *InMemoryDatabase {
	return &InMemoryDatabase{
		mutex:         sync.Mutex{},
		bundles:       make(map[uuid.UUID]*entity.Bundle),
		tokens:        make(map[uuid.UUID]*entity.JoinToken),
		trustDomains:  make(map[uuid.UUID]*entity.TrustDomain),
		relationships: make(map[uuid.UUID]*entity.Relationship),
	}
}

func (db *InMemoryDatabase) CreateOrUpdateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*entity.TrustDomain, error) {
	return nil, nil
}

func (db *InMemoryDatabase) DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error {
	return nil
}

func (db *InMemoryDatabase) ListTrustDomains(ctx context.Context) ([]*entity.TrustDomain, error) {
	return nil, nil
}

func (db *InMemoryDatabase) FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error) {
	return nil, nil
}

func (db *InMemoryDatabase) FindTrustDomainByName(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.TrustDomain, error) {
	return nil, nil
}

func (db *InMemoryDatabase) CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error) {
	return nil, nil
}

func (db *InMemoryDatabase) FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error) {
	return nil, nil
}

func (db *InMemoryDatabase) FindBundleByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) (*entity.Bundle, error) {
	return nil, nil
}

func (db *InMemoryDatabase) ListBundles(ctx context.Context) ([]*entity.Bundle, error) {
	return nil, nil
}

func (db *InMemoryDatabase) DeleteBundle(ctx context.Context, bundleID uuid.UUID) error {
	return nil
}

func (db *InMemoryDatabase) CreateJoinToken(ctx context.Context, req *entity.JoinToken) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.toFail {
		return nil, errors.New("fail to connect to in memory db")
	}

	req.ID = uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	req.Used = false
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.ExpiresAt = time.Now().Add(1 * time.Hour)

	db.tokens[req.ID.UUID] = req

	return req, nil
}

func (db *InMemoryDatabase) FindJoinTokensByID(ctx context.Context, joinTokenID uuid.UUID) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.toFail {
		return nil, errors.New("fail to connect to in memory db")
	}

	for id, jt := range db.tokens {
		if joinTokenID == id {
			return jt, nil
		}
	}
	return nil, nil
}

func (db *InMemoryDatabase) FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.JoinToken, error) {
	return nil, nil
}

func (db *InMemoryDatabase) ListJoinTokens(ctx context.Context) ([]*entity.JoinToken, error) {
	return nil, nil
}

func (db *InMemoryDatabase) UpdateJoinToken(ctx context.Context, joinTokenID uuid.UUID, used bool) (*entity.JoinToken, error) {
	return nil, nil
}

func (db *InMemoryDatabase) DeleteJoinToken(ctx context.Context, joinTokenID uuid.UUID) error {
	return nil
}

func (db *InMemoryDatabase) FindJoinToken(ctx context.Context, token string) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.toFail {
		return nil, errors.New("fail to connect to in memory db")
	}

	for _, jt := range db.tokens {
		if token == jt.Token {
			return jt, nil
		}
	}

	return nil, nil
}

func (db *InMemoryDatabase) CreateOrUpdateRelationship(ctx context.Context, req *entity.Relationship) (*entity.Relationship, error) {
	return nil, nil
}

func (db *InMemoryDatabase) FindRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error) {
	return nil, nil
}

func (db *InMemoryDatabase) FindRelationshipsByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.Relationship, error) {
	return nil, nil
}

func (db *InMemoryDatabase) ListRelationships(ctx context.Context) ([]*entity.Relationship, error) {
	return nil, nil
}

func (db *InMemoryDatabase) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	return nil
}

func (db *InMemoryDatabase) FailNext() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.toFail = true
}
