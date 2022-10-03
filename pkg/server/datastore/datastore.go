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
	GenerateAccessToken(ctx context.Context, t *common.AccessToken, trustDomain string) (*AccessToken, error)
	FetchAccessToken(ctx context.Context, token string) (*AccessToken, error)
}

type AccessToken struct {
	Token  string
	Expiry time.Time
	Member *Member
}

type Member struct {
	ID uuid.UUID

	Name        string
	TrustDomain string
}

type Relationship struct {
	ID uuid.UUID

	MemberA *Member
	MemberB *Member
}

// TODO: use until an actual DataStore implementation is added.

type MemStore struct {
	members      map[string]*Member // trust_domain (e.g. 'example.org') -> member
	relationship []*Relationship
	tokens       map[string]*AccessToken // token uuid string -> access token

	mu sync.RWMutex
}

func NewMemStore() DataStore {
	return &MemStore{
		members: make(map[string]*Member),
		tokens:  make(map[string]*AccessToken),
		mu:      sync.RWMutex{},
	}
}

func (s *MemStore) CreateMember(_ context.Context, member *common.Member) (*Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exist := s.members[member.TrustDomain.String()]; exist {
		return nil, fmt.Errorf("member already exists: %s", member.TrustDomain)
	}

	m := &Member{
		ID:          uuid.New(),
		Name:        member.Name,
		TrustDomain: member.TrustDomain.String(),
	}

	s.members[m.TrustDomain] = m

	return m, nil
}

func (s *MemStore) CreateRelationship(_ context.Context, rel *common.Relationship) (*Relationship, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rel.TrustDomainA.Compare(rel.TrustDomainB) == 0 {
		return nil, fmt.Errorf("cannot create relationship: trust domain members are the same: %s", rel.TrustDomainA)
	}

	if _, ok := s.members[rel.TrustDomainA.String()]; !ok {
		return nil, fmt.Errorf("member not found for trust domain: %s", rel.TrustDomainA)
	}
	if _, ok := s.members[rel.TrustDomainB.String()]; !ok {
		return nil, fmt.Errorf("member not found for trust domain: %s", rel.TrustDomainB)
	}
	r := &Relationship{
		ID:      uuid.New(),
		MemberA: s.members[rel.TrustDomainA.String()],
		MemberB: s.members[rel.TrustDomainB.String()],
	}

	s.relationship = append(s.relationship, r)

	return r, nil
}

func (s *MemStore) GenerateAccessToken(_ context.Context, token *common.AccessToken, trustDomain string) (*AccessToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	member := s.members[trustDomain]
	if member == nil {
		return nil, fmt.Errorf("failed to find member for the trust domain: %s", trustDomain)
	}

	at := &AccessToken{
		Token:  token.Token,
		Expiry: token.Expiry,
		Member: member,
	}

	s.tokens[at.Token] = at

	return at, nil
}

func (s *MemStore) FetchAccessToken(_ context.Context, token string) (*AccessToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	at, ok := s.tokens[token]
	if !ok {
		return nil, errors.New("failed to find token")
	}
	return at, nil
}
