package datastore

import (
	"context"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/google/uuid"
	"sync"
	"time"
)

type DataStore interface {
	CreateMember(ctx context.Context, member *common.Member) (*Member, error)
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
	member       []Member
	relationship []Relationship

	mu sync.RWMutex
}

func NewMemStore() DataStore {
	return &MemStore{
		mu: sync.RWMutex{},
	}
}

func (s *MemStore) CreateMember(_ context.Context, member *common.Member) (*Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens := make([]AccessToken, 0)
	for _, t := range member.Tokens {
		tokens = append(tokens, AccessToken{Token: t.Token})
	}
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	m := Member{
		ID:          id,
		Name:        member.Name,
		TrustDomain: member.TrustDomain,
		Tokens:      tokens,
	}
	s.member = append(s.member, m)
	fmt.Println("Members:", s.member)
	return &m, nil
}

func (s *MemStore) CreateRelationship(_ context.Context, rel *common.Relationship) (*Relationship, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.relationship = append(s.relationship, Relationship{
		MemberA: rel.MemberB,
		MemberB: rel.MemberB,
	})
	return &s.relationship[len(s.relationship)-1], nil
}

func (s *MemStore) CreateAccessToken(_ context.Context, t *common.AccessToken, memberID uuid.UUID) (*AccessToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i, err := s.getMemberIndex(memberID)
	if err != nil {
		return nil, err
	}

	s.member[i].Tokens = append(s.member[i].Tokens, AccessToken{Token: t.Token})
	// TODO: what to return here?
	return &s.member[i].Tokens[len(s.member[i].Tokens)-1], nil
}

func (s *MemStore) getMemberIndex(id uuid.UUID) (int, error) {
	for i, m := range s.member {
		if m.ID == id {
			return i, nil
		}
	}
	return 0, fmt.Errorf("member not found")
}
