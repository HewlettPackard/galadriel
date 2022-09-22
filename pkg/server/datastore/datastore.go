package datastore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/google/uuid"
)

type DataStore interface {
	CreateMember(ctx context.Context, m *common.Member) (*Member, error)
	CreateRelationship(ctx context.Context, r *common.Relationship) (*Relationship, error)
	CreateAccessToken(ctx context.Context, t *common.AccessToken, memberID uuid.UUID) (*AccessToken, error)
}

type AccessToken struct {
	Token  string
	Expiry time.Time
}

type Member struct {
	ID uuid.UUID

	Name        string
	TrustDomain string
	Tokens      []AccessToken
}

type Relationship struct {
	ID uuid.UUID

	MemberA uuid.UUID
	MemberB uuid.UUID
}

// TODO: use until an actual DataStore implementation is added.

type MemStore struct {
	member       map[uuid.UUID]*Member
	relationship map[uuid.UUID]*Relationship

	mu sync.RWMutex
}

func NewMemStore() DataStore {
	return &MemStore{
		member:       make(map[uuid.UUID]*Member),
		relationship: make(map[uuid.UUID]*Relationship),
		mu:           sync.RWMutex{},
	}
}

func (s *MemStore) CreateMember(_ context.Context, member *common.Member) (*Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m := Member{
		ID:          uuid.New(),
		Name:        member.Name,
		TrustDomain: member.TrustDomain,
	}

	for _, t := range member.Tokens {
		m.Tokens = append(m.Tokens, AccessToken{Token: t.Token, Expiry: t.Expiry})
	}
	s.member[m.ID] = &m

	fmt.Println("Members:", s.member)
	return s.member[m.ID], nil
}

func (s *MemStore) CreateRelationship(_ context.Context, rel *common.Relationship) (*Relationship, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r := Relationship{
		ID:      uuid.New(),
		MemberA: rel.MemberB,
		MemberB: rel.MemberB,
	}
	s.relationship[r.ID] = &r

	fmt.Println("Relationships:", s.relationship)
	return s.relationship[r.ID], nil
}

func (s *MemStore) CreateAccessToken(_ context.Context, t *common.AccessToken, memberID uuid.UUID) (*AccessToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.member[memberID]
	if !ok {
		return nil, errors.New("member not found")
	}

	m.Tokens = append(m.Tokens, AccessToken{Token: t.Token, Expiry: t.Expiry})
	s.member[memberID] = m

	return &m.Tokens[len(m.Tokens)-1], nil
}
