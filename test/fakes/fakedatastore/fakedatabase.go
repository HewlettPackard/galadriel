package fakedatastore

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

// WithRelationships overrides all relationships
func (db *FakeDatabase) WithRelationships(relationships ...*entity.Relationship) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.relationships = make(map[uuid.UUID]*entity.Relationship)
	for _, r := range relationships {
		db.relationships[r.ID.UUID] = r
	}
}

// WithTrustDomains overrides all trust domains
func (db *FakeDatabase) WithTrustDomains(trustDomains ...*entity.TrustDomain) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.trustDomains = make(map[uuid.UUID]*entity.TrustDomain)
	for _, td := range trustDomains {
		db.trustDomains[td.ID.UUID] = td
	}
}

// WithBundles overrides all bundles
func (db *FakeDatabase) WithBundles(bundles ...*entity.Bundle) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.bundles = make(map[uuid.UUID]*entity.Bundle)
	for _, b := range bundles {
		db.bundles[b.ID.UUID] = b
	}
}

// WithTokens overrides all tokens
func (db *FakeDatabase) WithTokens(bundles ...*entity.JoinToken) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.tokens = make(map[uuid.UUID]*entity.JoinToken)
	for _, jt := range bundles {
		db.tokens[jt.ID.UUID] = jt
	}
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

func (db *FakeDatabase) CreateOrUpdateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*entity.TrustDomain, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	if !req.ID.Valid {
		req.ID = uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		}
		req.CreatedAt = time.Now()
	}

	req.UpdatedAt = time.Now()
	db.trustDomains[req.ID.UUID] = req

	return req, nil
}

func (db *FakeDatabase) DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return err
	}

	var uuid uuid.UUID

	for idx, td := range db.trustDomains {
		if trustDomainID.String() == td.ID.UUID.String() {
			uuid = idx
		}
	}

	delete(db.trustDomains, uuid)

	return nil
}

func (db *FakeDatabase) ListTrustDomains(ctx context.Context) ([]*entity.TrustDomain, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	domains := []*entity.TrustDomain{}
	for _, td := range db.trustDomains {
		domains = append(domains, td)
	}

	return domains, nil
}

func (db *FakeDatabase) FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

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
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, td := range db.trustDomains {
		if trustDomain.String() == td.Name.String() {
			return td, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	if !req.ID.Valid {
		// it's an insert
		req.ID = uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		}
	}

	req.UpdatedAt = time.Now()
	db.bundles[req.ID.UUID] = req

	return req, nil
}

func (db *FakeDatabase) FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, b := range db.bundles {
		if bundleID.String() == b.ID.UUID.String() {
			return b, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) FindBundleByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) (*entity.Bundle, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, bundle := range db.bundles {
		if trustDomainID == bundle.TrustDomainID {
			return bundle, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) ListBundles(ctx context.Context) ([]*entity.Bundle, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	bundles := []*entity.Bundle{}
	for _, bundle := range db.bundles {
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func (db *FakeDatabase) DeleteBundle(ctx context.Context, bundleID uuid.UUID) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return err
	}

	var uuid uuid.UUID
	for idx, b := range db.bundles {
		if bundleID.String() == b.ID.UUID.String() {
			uuid = idx
		}
	}

	delete(db.bundles, uuid)

	return nil
}

func (db *FakeDatabase) CreateJoinToken(ctx context.Context, req *entity.JoinToken) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

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
		if joinTokenID.String() == id.String() {
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

	tokens := []*entity.JoinToken{}
	for _, jt := range db.tokens {
		if jt.TrustDomainID.String() == trustDomainID.String() {
			tokens = append(tokens, jt)
		}
	}

	return tokens, nil
}

func (db *FakeDatabase) ListJoinTokens(ctx context.Context) ([]*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	tokens := []*entity.JoinToken{}
	for _, jt := range db.tokens {
		tokens = append(tokens, jt)
	}

	return tokens, nil
}

func (db *FakeDatabase) UpdateJoinToken(ctx context.Context, joinTokenID uuid.UUID, used bool) (*entity.JoinToken, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, jt := range db.tokens {
		if jt.ID.UUID == joinTokenID {
			jt.Used = used
			jt.UpdatedAt = time.Now()
			return jt, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) DeleteJoinToken(ctx context.Context, joinTokenID uuid.UUID) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return err
	}

	var uuid uuid.UUID
	for idx, t := range db.tokens {
		if t.ID.UUID.String() == joinTokenID.String() {
			uuid = idx
		}
	}

	delete(db.tokens, uuid)

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
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	if !req.ID.Valid {
		req.ID = uuid.NullUUID{
			Valid: true,
			UUID:  uuid.New(),
		}
		req.CreatedAt = time.Now()
	}

	req.UpdatedAt = time.Now()
	db.relationships[req.ID.UUID] = req

	return req, nil
}

func (db *FakeDatabase) FindRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	for _, r := range db.relationships {
		if r.ID.UUID.String() == relationshipID.String() {
			return r, nil
		}
	}

	return nil, nil
}

func (db *FakeDatabase) FindRelationshipsByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.Relationship, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	var relationships []*entity.Relationship
	for _, r := range db.relationships {
		matchA := r.TrustDomainAID.String() == trustDomainID.String()
		matchB := r.TrustDomainBID.String() == trustDomainID.String()

		if matchA || matchB {
			relationships = append(relationships, r)
		}
	}

	return relationships, nil
}

func (db *FakeDatabase) ListRelationships(ctx context.Context) ([]*entity.Relationship, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return nil, err
	}

	var relationships []*entity.Relationship
	for _, r := range db.relationships {
		relationships = append(relationships, r)
	}

	return relationships, nil
}

func (db *FakeDatabase) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.getNextError(); err != nil {
		return err
	}

	var uuid uuid.UUID
	for idx, r := range db.relationships {
		if r.ID.UUID.String() == relationshipID.String() {
			uuid = idx
		}
	}

	delete(db.relationships, uuid)

	return nil
}
