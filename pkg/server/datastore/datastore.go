package datastore

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/google/uuid"
)

type DataStore interface {
	CreateMember(ctx context.Context, m *common.Member) (*Member, error)
	CreateRelationship(ctx context.Context, r *common.Relationship) (*Relationship, error)
	GenerateAccessToken(ctx context.Context, t *common.AccessToken, trustDomain string) (*AccessToken, error)
}

type AccessToken struct {
	Token    string
	Expiry   time.Time
	MemberID uuid.UUID
}

type Member struct {
	ID uuid.UUID

	Name        string
	TrustDomain string
}

type Relationship struct {
	ID uuid.UUID

	TrustDomainA string
	TrustDomainB string
}

// TODO: use until an actual DataStore implementation is added.

type MemStore struct {
	member       map[string]*Member
	relationship []*Relationship
	token        []*AccessToken

	mu sync.RWMutex
}

func NewMemStore() DataStore {
	return &MemStore{
		member: make(map[string]*Member),
		mu:     sync.RWMutex{},
	}
}

func (s *MemStore) CreateMember(_ context.Context, member *common.Member) (*Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m := &Member{
		ID:          uuid.New(),
		Name:        member.Name,
		TrustDomain: member.TrustDomain,
	}

	s.member[m.TrustDomain] = m

	return m, nil
}

func (s *MemStore) CreateRelationship(_ context.Context, rel *common.Relationship) (*Relationship, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.member[rel.TrustDomainA]; !ok {
		return nil, errors.New("member not found")
	}
	if _, ok := s.member[rel.TrustDomainB]; !ok {
		return nil, errors.New("member not found")
	}
	r := &Relationship{
		ID:           uuid.New(),
		TrustDomainA: rel.TrustDomainA,
		TrustDomainB: rel.TrustDomainB,
	}

	s.relationship = append(s.relationship, r)

	return r, nil
}

func (s *MemStore) GenerateAccessToken(_ context.Context, token *common.AccessToken, td string) (*AccessToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var memberID uuid.UUID
	for _, member := range s.member {
		if member.TrustDomain == td {
			memberID = member.ID
		}
	}

	at := &AccessToken{
		Token:    token.Token,
		Expiry:   token.Expiry,
		MemberID: memberID,
	}

	s.token = append(s.token, at)

	return at, nil
}
