package session

import (
	"context"
	"sync"
)

type memoryStore struct {
	data map[string]SessionData
	sync.RWMutex
}

func (s *memoryStore) Get(ctx context.Context, id string) (SessionData, error) {
	s.RLock()
	defer s.RUnlock()
	data, ok := s.data[id]
	if !ok {
		return SessionData{}, ErrSessionNotFound
	}
	return data, nil
}

func (s *memoryStore) Set(ctx context.Context, session SessionData) error {
	s.Lock()
	defer s.Unlock()

	dataCopy := make(map[string]any, len(session.Data))
	for k, v := range session.Data {
		dataCopy[k] = v
	}

	cpy := SessionData{
		ID:            session.ID,
		Data:          dataCopy,
		CreatedAt:     session.CreatedAt,
		ExpiresAt:     session.ExpiresAt,
		UpdatedAt:     session.UpdatedAt,
		Authenticated: session.Authenticated,
		UserID:        session.UserID,
	}
	s.data[session.ID] = cpy

	return nil
}

func (s *memoryStore) Delete(ctx context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	delete(s.data, id)
	return nil
}

func NewMemoryStore() Store {
	return &memoryStore{
		data: make(map[string]SessionData),
	}
}
