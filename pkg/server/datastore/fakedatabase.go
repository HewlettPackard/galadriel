package datastore

import (
	"context"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type FakeDatabase struct {
	mutex  sync.Mutex
	errors []error

	// Entities
	bundles       map[uuid.UUID]*entity.Bundle
	tokens        map[uuid.UUID]*entity.JoinToken
	trustDomains  map[uuid.UUID]*entity.TrustDomain
	relationships map[uuid.UUID]*entity.Relationship
}

func NewFakeDB() *FakeDatabase {

	return &FakeDatabase{
		errors: []error{},
		mutex:  sync.Mutex{},

		bundles:       make(map[uuid.UUID]*entity.Bundle),
		tokens:        make(map[uuid.UUID]*entity.JoinToken),
		trustDomains:  make(map[uuid.UUID]*entity.TrustDomain),
		relationships: make(map[uuid.UUID]*entity.Relationship),
	}
}

func (db *FakeDatabase) CreateOrUpdateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*entity.TrustDomain, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	req.ID = uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	db.trustDomains[req.ID.UUID] = req

	return req, nil
}

func (db *FakeDatabase) DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error {
	if err := db.getNextError(); err != nil {
		return err
	}

	return nil
}

func (db *FakeDatabase) ListTrustDomains(ctx context.Context) ([]*entity.TrustDomain, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, td := range db.trustDomains {
		if trustDomainID.String() == td.ID.UUID.String() {
			return td, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) FindTrustDomainByName(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.TrustDomain, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	req.ID = uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	db.bundles[req.ID.UUID] = req

	return req, nil
}

func (db *FakeDatabase) FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) FindBundleByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) (*entity.Bundle, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, bundle := range db.bundles {
		if trustDomainID.String() == bundle.TrustDomainID.String() {
			return bundle, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) ListBundles(ctx context.Context) ([]*entity.Bundle, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) DeleteBundle(ctx context.Context, bundleID uuid.UUID) error {
	if err := db.getNextError(); err != nil {
		return err
	}

	return nil
}

func (db *FakeDatabase) CreateJoinToken(ctx context.Context, req *entity.JoinToken) (*entity.JoinToken, error) {

	if err := db.getNextError(); err != nil {
		return nil, err
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

func (db *FakeDatabase) FindJoinTokensByID(ctx context.Context, joinTokenID uuid.UUID) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for id, jt := range db.tokens {
		if joinTokenID == id {
			return jt, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) ListJoinTokens(ctx context.Context) ([]*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) UpdateJoinToken(ctx context.Context, joinTokenID uuid.UUID, used bool) (*entity.JoinToken, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) DeleteJoinToken(ctx context.Context, joinTokenID uuid.UUID) error {
	if err := db.getNextError(); err != nil {
		return err
	}

	return nil
}

func (db *FakeDatabase) FindJoinToken(ctx context.Context, token string) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, jt := range db.tokens {
		if token == jt.Token {
			return jt, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) CreateOrUpdateRelationship(ctx context.Context, req *entity.Relationship) (*entity.Relationship, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) FindRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) FindRelationshipsByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.Relationship, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) ListRelationships(ctx context.Context) ([]*entity.Relationship, error) {
	if err := db.getNextError(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *FakeDatabase) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	if err := db.getNextError(); err != nil {
		return err
	}

	return nil
}

func (db *FakeDatabase) SetNextError(err error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.errors = []error{err}
}

func (db *FakeDatabase) AppendNextError(err error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.errors = append(db.errors, err)
}

func (db *FakeDatabase) getNextError() error {
	if len(db.errors) == 0 {
		return nil
	}

	err := db.errors[0]
	db.errors = db.errors[1:]

	return err
}
