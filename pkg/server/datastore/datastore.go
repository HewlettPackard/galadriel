package datastore

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DataStore interface {
	CreateJoinToken(context.Context, *JoinToken) error
	DeleteJoinToken(ctx context.Context, token string) error
	FetchJoinToken(ctx context.Context, token string) (*JoinToken, error)
}

type JoinToken struct {
	Token  string
	Expiry time.Time
}

// TODO: use until an actual DataStore implementation is added.
type MemStore struct {
	mu     sync.RWMutex
	tokens []*JoinToken
}

func NewMemStore() DataStore {
	return &MemStore{
		mu:     sync.RWMutex{},
		tokens: nil,
	}
}

func (s *MemStore) CreateJoinToken(ctx context.Context, token *JoinToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens = append(s.tokens, token)
	return nil
}

func (s *MemStore) FetchJoinToken(ctx context.Context, token string) (*JoinToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tokens {
		if t.Token == token {
			return t, nil
		}

	}
	return nil, fmt.Errorf("token not found")
}

func (s *MemStore) DeleteJoinToken(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	index, err := s.getIndex(token)
	if err != nil {
		return err
	}
	s.tokens = removeIndex(s.tokens, index)
	return nil
}

func (s *MemStore) getIndex(token string) (int, error) {
	for i, t := range s.tokens {
		if t.Token == token {
			return i, nil
		}
	}
	return 0, fmt.Errorf("token not found")
}

func removeIndex(s []*JoinToken, index int) []*JoinToken {
	return append(s[:index], s[index+1:]...)
}
